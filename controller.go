package main

import (
	travisclientset "github.com/travis-ci/trvs-operator/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

func NewController(kubeclient kubernetes.Interface, travisclient travisclientset.Interface) *Controller {
	return &Controller{
		kubeclient:   kubeclient,
		travisclient: travisclient,
	}
}

type Controller struct {
	kubeclient   kubernetes.Interface
	travisclient travisclientset.Interface
}

func (c *Controller) Run(threads int, stopCh <-chan struct{}) error {
	return nil
}
