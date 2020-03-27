package controller

import (
	"context"
	"fmt"
	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-recovery/pkg/sts"
	log "github.com/sirupsen/logrus"
	k8s_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	uruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

type Controller struct {
	cfg        *rest.Config
	queue      workqueue.RateLimitingInterface
	informer   cache.SharedIndexInformer
	maxRetries int
}

func NewController(cfg *rest.Config) *Controller {
	return &Controller{
		cfg: cfg,
	}
}

func (c *Controller) Run(ctx context.Context) {
	log.Infoln("Starting the controller...")
	clientSet, err := kubernetes.NewForConfig(c.cfg)
	if err == nil {
		stopCh := make(chan struct{})
		defer close(stopCh)
		c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		c.informer = cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return clientSet.CoreV1().Events(meta_v1.NamespaceAll).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return clientSet.CoreV1().Events(meta_v1.NamespaceAll).Watch(options)
				},
			},
			&k8s_v1.Event{},
			0, //No resync
			cache.Indexers{},
		)

		c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					c.queue.Add(key)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldEvent, _ := old.(*k8s_v1.Event)
				newEvent, _ := new.(*k8s_v1.Event)
				if oldEvent.ResourceVersion != newEvent.ResourceVersion {
					key, err := cache.MetaNamespaceKeyFunc(new)
					if err == nil {
						c.queue.Add(key)
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err == nil {
					c.queue.Add(key)
				}
			},
		})

		go c.informer.Run(stopCh)
		log.Infoln("Controller started.")
		if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
			uruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return
		}
		log.Infoln("Controller synced.")

		wait.Until(c.runWorker, time.Second, stopCh)
	} else {
		log.Fatalf("Cannot create the kubernetes client: %v", err)
	}

}

func (c *Controller) runWorker() {
	for c.processNext() {
	}
}

func (c *Controller) processNext() bool {
	key, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	if err == nil {
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < c.maxRetries {
		log.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		log.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		uruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	obj, _, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}
	sts.Process(obj)
	return nil
}
