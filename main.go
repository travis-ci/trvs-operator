package main

import (
	log "github.com/sirupsen/logrus"
	travisclientset "github.com/travis-ci/trvs-operator/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	stopCh := make(chan struct{})
	log.SetLevel(log.DebugLevel)

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

	controller := NewController(kubeclient, travisclient)

	if err := controller.Run(2, stopCh); err != nil {
		log.WithError(err).Fatal("error running controller")
	}
}
