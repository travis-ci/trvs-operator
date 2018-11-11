package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	travisclientset "github.com/travis-ci/trvs-operator/pkg/client/clientset/versioned"
	travisscheme "github.com/travis-ci/trvs-operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/travis-ci/trvs-operator/pkg/client/informers/externalversions/travisci/v1"
	listers "github.com/travis-ci/trvs-operator/pkg/client/listers/travisci/v1"
)

func NewController(
	keychains Keychains,
	kubeclient kubernetes.Interface,
	travisclient travisclientset.Interface,
	trvsSecretInformer informers.TrvsSecretInformer) *Controller {

	runtime.Must(travisscheme.AddToScheme(scheme.Scheme))
	controller := &Controller{
		keychains:    keychains,
		kubeclient:   kubeclient,
		travisclient: travisclient,
		trvsLister:   trvsSecretInformer.Lister(),
		trvsSynced:   trvsSecretInformer.Informer().HasSynced,
		workqueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TrvsSecrets"),
	}

	trvsSecretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueTrvsSecret,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueTrvsSecret(new)
		},
	})

	return controller
}

type Controller struct {
	keychains Keychains

	kubeclient   kubernetes.Interface
	travisclient travisclientset.Interface

	secretsLister corelisters.SecretLister
	trvsLister    listers.TrvsSecretLister
	trvsSynced    cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
}

func (c *Controller) Run(threads int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	log.Info("starting controller")

	log.Info("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.trvsSynced); !ok {
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

	trvsSecret, err := c.trvsLister.TrvsSecrets(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			entry.Error("resource no longer exists")
			return nil
		}

		return err
	}

	entry.WithFields(log.Fields{
		"name": trvsSecret.ObjectMeta.Name,
		"app":  trvsSecret.Spec.App,
		"env":  trvsSecret.Spec.Environment,
	}).Info("checking secret")

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
