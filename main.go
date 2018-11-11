package main

import (
	log "github.com/sirupsen/logrus"
	travisclientset "github.com/travis-ci/trvs-operator/pkg/client/clientset/versioned"
	informers "github.com/travis-ci/trvs-operator/pkg/client/informers/externalversions"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	stopCh := setupSignalHandler()
	log.SetLevel(log.DebugLevel)

	keychains := setupKeychains()
	keychains.Watch(30 * time.Second)

	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.WithError(err).Fatal("could not build config")
	}

	kubeclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("could not create kubernetes client")
	}

	travisclient, err := travisclientset.NewForConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("could not create custom client")
	}

	travisInformerFactory := informers.NewSharedInformerFactory(travisclient, time.Second*30)

	controller := NewController(keychains, kubeclient, travisclient,
		travisInformerFactory.Travisci().V1().TrvsSecrets())

	travisInformerFactory.Start(stopCh)

	if err := controller.Run(2, stopCh); err != nil {
		log.WithError(err).Fatal("error running controller")
	}
}

func setupSignalHandler() <-chan struct{} {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}

func setupKeychains() Keychains {
	var ks Keychains

	ks.Org = createKeychain("travis-keychain")
	ks.Com = createKeychain("travis-pro-keychain")

	return ks
}

func createKeychain(name string) *Keychain {
	entry := log.WithField("name", name)

	urlFile := "/etc/secrets/" + name + "-url"
	keyFile := "/etc/secrets/" + name + ".key"

	url, err := ioutil.ReadFile(urlFile)
	if err != nil {
		entry.WithError(err).WithField("file", urlFile).Fatal("could not read url file")
	}

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		entry.WithError(err).WithField("file", keyFile).Fatal("could not read key file")
	}

	k, err := NewKeychain(name, string(url), key)
	if err != nil {
		entry.WithError(err).Fatal("could not create keychain")
	}

	return k
}
