package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	var kubeconfig string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file")
	flag.Set("logtostderr", "true")
	flag.Parse()

	LogInit(false)

	client, err := k8sGetClient(kubeconfig)
	if err != nil {
		glog.Error(fmt.Errorf("Failed to get client: %v", err))
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	controller := newServiceController(client, metav1.NamespaceAll, 10*time.Second)
	controller.Run(stopCh)

}
