package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	uruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-recovery/pkg/sts"
)

const (
	InstanceLabel = "kudo.dev/instance"
)

type Controller struct {
	client     *kubernetes.Clientset
	queue      workqueue.RateLimitingInterface
	informer   cache.SharedIndexInformer
	maxRetries int
}

func NewController(client *kubernetes.Clientset) *Controller {
	return &Controller{
		client: client,
	}
}

func (c *Controller) Run(ctx context.Context) {
	namespace := os.Getenv("NAMESPACE")
	instance := os.Getenv("INSTANCE_NAME")

	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	log.Infof("Starting the controller for namespace %s...", namespace)

	labelSelector := ""
	if instance != "" {
		labelSelector = fmt.Sprintf("%s = %s", InstanceLabel, instance)
		log.Infof("Acting only on pods with KUDO instance label %s", instance)
	} else {
		log.Infof("Acting on ALL pods in selected namespace")
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	c.informer = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = labelSelector
				return c.client.CoreV1().Pods(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = labelSelector
				return c.client.CoreV1().Pods(namespace).Watch(options)
			},
		},
		&corev1.Pod{},
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
			oldEvent, _ := old.(*corev1.Pod)
			newEvent, _ := new.(*corev1.Pod)
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
		uruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	log.Infoln("Controller synced.")

	wait.Until(c.runWorker, time.Second, stopCh)
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
	if obj == nil {
		return nil
	}
	sts.Process(c.client, obj)
	return nil
}
