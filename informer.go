package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	lister_v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func k8sGetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func k8sGetClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := k8sGetClientConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Construct the Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type serviceController struct {
	client        kubernetes.Interface
	informer      cache.Controller
	indexer       cache.Indexer
	serviceLister lister_v1.ServiceLister
	updater       *Updater
}

func newServiceController(client kubernetes.Interface, namespace string, updateInterval time.Duration) *serviceController {
	sc := &serviceController{
		client: client,
	}
	sc.updater = newUpdater()

	indexer, informer := cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (runtime.Object, error) {
				return client.Core().Services(namespace).List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return client.Core().Services(namespace).Watch(lo)

			},
		},
		// The types of objects this informer will return
		&v1.Service{},
		// The resync period of this object. This will force a re-queue of all cached objects at this interval.
		// Every object will trigger the `Updatefunc` even if there have been no actual updates triggered.
		// In some cases you can set this to a very high interval - as you can assume you will see periodic
		// updates in normal operation.
		updateInterval,
		// Callback Functions to trigger on add/update/delete
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
					log.Debugf("addFunc key: %v", key)
					service := obj.(*v1.Service)
					event, err := newUpdaterEvent(eventCreate, service)
					if err != nil {
						log.Debugf("failed to generate new updater event: %v", err.Error())
					} else {
						sc.updater.events <- *event
					}

				}
			},
			UpdateFunc: func(old, new interface{}) {
				if key, err := cache.MetaNamespaceKeyFunc(new); err == nil {
					log.Debugf("updateFunc key: %v", key)
					service := new.(*v1.Service) // TODO: do we need to do a diff on old/new?
					event, err := newUpdaterEvent(eventUpdate, service)
					if err != nil {
						log.Debugf("failed to generate new updater event: %v", err.Error())
					} else {
						sc.updater.events <- *event
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err == nil {
					log.Debugf("deleteFunc key: %v", key)
					service := obj.(*v1.Service)
					event, err := newUpdaterEvent(eventDelete, service)
					if err != nil {
						log.Debugf("failed to generate new updater event: %v", err.Error())
					} else {
						sc.updater.events <- *event
					}
				}
			},
		},
		cache.Indexers{},
	)

	sc.informer = informer
	sc.indexer = indexer
	sc.serviceLister = lister_v1.NewServiceLister(indexer)

	return sc
}

func (c *serviceController) Run(stopCh chan struct{}) {
	log.Info("Starting serviceController")

	go c.informer.Run(stopCh)
	go c.updater.Run(stopCh)

	<-stopCh
	log.Info("Stopping serviceController")
}
