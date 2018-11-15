package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	travisv1 "github.com/travis-ci/trvs-operator/pkg/apis/travisci/v1"
	travisclientset "github.com/travis-ci/trvs-operator/pkg/client/clientset/versioned"
	travisscheme "github.com/travis-ci/trvs-operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/travis-ci/trvs-operator/pkg/client/informers/externalversions/travisci/v1"
	listers "github.com/travis-ci/trvs-operator/pkg/client/listers/travisci/v1"
)

const controllerAgentName = "trvs-operator"

const (
	ErrResourceExists     = "ErrResourceExists"
	MessageResourceExists = "Secret %q already exists and is not managed by a TrvsSecret"
)

func NewController(
	keychains Keychains,
	keychainSyncPeriod time.Duration,
	kubeclient kubernetes.Interface,
	travisclient travisclientset.Interface,
	secretInformer coreinformers.SecretInformer,
	trvsSecretInformer informers.TrvsSecretInformer) *Controller {

	runtime.Must(travisscheme.AddToScheme(scheme.Scheme))
	log.Info("creating event recorder")
	eb := record.NewBroadcaster()
	eb.StartLogging(log.WithField("type", "events").Infof)
	eb.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: kubeclient.CoreV1().Events(""),
	})

	recorder := eb.NewRecorder(scheme.Scheme, v1.EventSource{
		Component: controllerAgentName,
	})

	controller := &Controller{
		keychains:     keychains,
		kubeclient:    kubeclient,
		travisclient:  travisclient,
		secretsLister: secretInformer.Lister(),
		secretsSynced: secretInformer.Informer().HasSynced,
		trvsLister:    trvsSecretInformer.Lister(),
		trvsSynced:    trvsSecretInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TrvsSecrets"),
		recorder:      recorder,
	}

	keychains.Watch(keychainSyncPeriod, controller.enqueueKeychainSecrets)

	trvsSecretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueTrvsSecret,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueTrvsSecret(new)
		},
	})

	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			oldSec := old.(*v1.Secret)
			newSec := new.(*v1.Secret)
			if newSec.ResourceVersion == oldSec.ResourceVersion {
				// Ignore the update if the secret hasn't actually changed.
				//
				// This is needed because the UpdateFunc is called periodically even when there
				// are no changes, so updates aren't missed. We already get those updates from the
				// other informer, though, so this is redundant.
				return
			}

			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

type Controller struct {
	keychains Keychains

	kubeclient   kubernetes.Interface
	travisclient travisclientset.Interface

	secretsLister corelisters.SecretLister
	secretsSynced cache.InformerSynced
	trvsLister    listers.TrvsSecretLister
	trvsSynced    cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

func (c *Controller) Run(threads int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	log.Info("starting controller")

	log.Info("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.secretsSynced, c.trvsSynced); !ok {
		return fmt.Errorf("failed waiting for caches to sync")
	}

	entry := log.WithField("count", threads)
	entry.Info("starting workers")

	for i := 0; i < threads; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	entry.Info("started workers")
	<-stopCh
	entry.Info("stopping workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	func(obj interface{}) {
		defer c.workqueue.Done(obj)

		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			log.WithField("value", obj).Error("unexpected value in workqueue")
			return
		}

		entry := log.WithField("key", key)

		entry.Info("got workqueue item")
		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			entry.WithError(err).Error("could not process item")
			return
		}

		c.workqueue.Forget(obj)
		entry.Info("synced secret")
	}(obj)

	return true
}

func (c *Controller) syncHandler(key string) error {
	entry := log.WithField("key", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		entry.Error("invalid resource key")
		return nil
	}

	ts, err := c.trvsLister.TrvsSecrets(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			entry.Error("resource no longer exists")
			return nil
		}

		return err
	}

	entry = entry.WithFields(log.Fields{
		"namespace": ts.Namespace,
		"name":      ts.Name,
		"app":       ts.Spec.App,
		"env":       ts.Spec.Environment,
	})
	entry.Info("checking secret")

	secretValues, err := trvs.Generate(ts.Spec)
	if err != nil {
		entry.WithError(err).Error("could not get secret data from keychain")
		return nil
	}

	entry.WithField("keys", len(secretValues)).Info("found secret data in keychain")

	secret, err := c.secretsLister.Secrets(ts.Namespace).Get(ts.Name)
	if errors.IsNotFound(err) {
		secret, err = c.kubeclient.CoreV1().Secrets(ts.Namespace).Create(newSecret(ts, secretValues))
		if err == nil {
			c.recorder.Eventf(ts, v1.EventTypeNormal, "CreateSecret", "Created secret: %s", secret.Name)
		}
	}

	if err != nil {
		entry.WithError(err).Error("could not find/create secret")
		return nil
	}

	if !metav1.IsControlledBy(secret, ts) {
		msg := fmt.Sprintf(MessageResourceExists, secret.Name)
		c.recorder.Event(ts, v1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	if reflect.DeepEqual(secretValues, secret.Data) {
		entry.Info("secret is already up-to-date")
		return nil
	}

	entry.Info("updating secret")
	secret, err = c.kubeclient.CoreV1().Secrets(ts.Namespace).Update(newSecret(ts, secretValues))
	if err != nil {
		return err
	}

	c.recorder.Eventf(ts, v1.EventTypeNormal, "UpdateSecret", "Updated secret: %s", secret.Name)
	return nil
}

func (c *Controller) enqueueTrvsSecret(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool

	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			log.Error("could not decode secret object")
			return
		}

		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			log.Error("could not decode tombstone object")
		}

		log.Info("recovered secret from tombstone")
	}

	log.WithField("name", object.GetName()).Info("finding original of secret")
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "TrvsSecret" {
			return
		}

		ts, err := c.trvsLister.TrvsSecrets(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			log.WithFields(log.Fields{
				"self":  object.GetSelfLink(),
				"owner": ownerRef.Name,
			}).Warn("ignoring orphaned secret")
		}

		c.enqueueTrvsSecret(ts)
	}
}

func (c *Controller) enqueueKeychainSecrets(k *Keychain) {
	isPro := k.IsPro()

	secrets, err := c.trvsLister.List(labels.Everything())
	if err != nil {
		log.WithError(err).Error("could not fetch existing secrets")
		return
	}

	for _, ts := range secrets {
		// if the secret matches this keychain, enqueue it so we check for updates
		if ts.Spec.IsPro == isPro {
			c.enqueueTrvsSecret(ts)
		}
	}
}

func newSecret(ts *travisv1.TrvsSecret, data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ts.Name,
			Namespace: ts.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ts, schema.GroupVersionKind{
					Group:   travisv1.SchemeGroupVersion.Group,
					Version: travisv1.SchemeGroupVersion.Version,
					Kind:    "TrvsSecret",
				}),
			},
		},
		Data: data,
	}
}
