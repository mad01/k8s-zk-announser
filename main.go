package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	var kubeconfig string
	var debug bool
	var zookeeperAddr string
	var updateInterval time.Duration

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file")
	flag.StringVar(&zookeeperAddr, "zookeeper.addr", "localhost:2181", "zookeeper address:port")
	flag.DurationVar(&updateInterval, "interval", 10*time.Second, "interavl to update the informer cache")
	flag.BoolVar(&debug, "debug", false, "debug logging")
	flag.Set("logtostderr", "true")
	flag.Parse()

	LogInit(debug)

	client, err := k8sGetClient(kubeconfig)
	if err != nil {
		log.Error(fmt.Errorf("Failed to get client: %v", err))
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	controller := newServiceController(client, metav1.NamespaceAll, updateInterval, zookeeperAddr)
	controller.Run(stopCh)

}
