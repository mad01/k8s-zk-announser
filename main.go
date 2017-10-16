package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	lister_v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
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
	queue         workqueue.RateLimitingInterface
	serviceLister lister_v1.ServiceLister
}

func newServiceController(client kubernetes.Interface, namespace string, updateInterval time.Duration) *serviceController {
	sc := &serviceController{
		client: client,
		queue:  workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

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
		&v1.ServiceList{},
		// The resync period of this object. This will force a re-queue of all cached objects at this interval.
		// Every object will trigger the `Updatefunc` even if there have been no actual updates triggered.
		// In some cases you can set this to a very high interval - as you can assume you will see periodic
		// updates in normal operation.
		updateInterval,
		// Callback Functions to trigger on add/update/delete
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
					sc.queue.Add(key)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				if key, err := cache.MetaNamespaceKeyFunc(new); err == nil {
					sc.queue.Add(key)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err == nil {
					sc.queue.Add(key)
				}
			},
		},
		cache.Indexers{},
	)

	sc.informer = informer
	sc.serviceLister = lister_v1.NewServiceLister(indexer)

	return sc
}

func (c *serviceController) Run(stopCh chan struct{}) {
	defer c.queue.ShutDown()
	glog.Info("Starting serviceController")

	go c.informer.Run(stopCh)

	// Wait for all caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		glog.Error(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	// Launching additional goroutines would parallelize workers consuming from the queue (but we don't really need this)
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
	glog.Info("Stopping serviceController")
}

func (c *serviceController) runWorker() {
	for c.processNext() {
	}
}

func (c *serviceController) processNext() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)
	// Invoke the method containing the business logic
	err := c.process(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

func (c *serviceController) process(key string) error {
	selector := labels.NewSelector()

	services, err := c.serviceLister.List(selector)
	if err != nil {
		return fmt.Errorf("failed to retrieve node by key %q: %v", key, err)
	}

	for _, service := range services {
		fmt.Printf("%v %v\n", service.GetName(), service.GetNamespace())
	}

	return nil
}

func (c *serviceController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		glog.Infof("Error processing %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	glog.Errorf("Dropping node %q out of the queue: %v", key, err)
}

func getK8sServerVersion(client *kubernetes.Clientset) (string, error) {
	var err error
	if version, err := client.ServerVersion(); err == nil {
		return version.String(), nil
	}
	return "", err

}

func main() {
	var kubeconfig string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file")
	flag.Set("logtostderr", "true")
	flag.Parse()

	client, err := k8sGetClient(kubeconfig)
	if err != nil {
		glog.Error(fmt.Errorf("Failed to get clinet: %v", err))
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	controller := newServiceController(client, metav1.NamespaceAll, 10*time.Second)
	controller.Run(stopCh)

}
