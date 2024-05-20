package main

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	applisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	// ClientSet is the kubernetes client
	// Use when you want to know the current exact state of the resources
	ClientSet kubernetes.Interface

	// If you want to know the state of the resources in the most efficient way,
	// use this lister.
	//
	// Objects read from listers can always be slightly out-of-date (i.e., stale)
	// because the client has to first observe changes to API objects via watch events
	// and then update the cache.
	//
	// Thus, don’t make any decisions based on data read from Listers
	// if the consequences of deciding wrongfully based on stale state
	// might be catastrophic (e.g. leaking infrastructure resources).
	// In such cases, read directly from the API server via a client instead.
	//
	// Objects retrieved from Informers or Listers are pointers to the cached objects,
	// so they must not be modified without copying them first.
	//
	// Ref: https://medium.com/@timebertt/kubernetes-controllers-at-scale-clients-caches-conflicts-patches-explained-aa0f7a8b4332
	DeploymentLister applisters.DeploymentLister

	// Notifies the controller when the cache is synced
	DeploymentCacheSync cache.InformerSynced

	// Every time a new event detected by informer, it will be added to the queue
	Queue workqueue.RateLimitingInterface
}

func NewController(
	clientSet kubernetes.Interface,
	informer appinformers.DeploymentInformer,
) *Controller {
	c := &Controller{
		ClientSet:           clientSet,
		DeploymentLister:    informer.Lister(),
		DeploymentCacheSync: informer.Informer().HasSynced,
		Queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"example",
		),
	}

	informer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDelete,
		},
	)

	return c
}

func (c *Controller) Run(numWorkers int, stopCh <-chan struct{}) {
	defer func() {
		fmt.Println("Cleaning up...")
		utilruntime.HandleCrash()
		c.Queue.ShutDown()
		fmt.Println("Stopped")
	}()

	// Wait for the caches to be synced before starting workers
	if !cache.WaitForCacheSync(stopCh, c.DeploymentCacheSync) {
		fmt.Println("Error syncing cache")
	}

	for i := 0; i < numWorkers; i++ {
		// Wait every 1 second to process the next item in the queue
		go wait.Until(c.worker, 1*time.Second, stopCh)
	}

	// Block the main thread
	fmt.Println("Starting controller")
	<-stopCh
	fmt.Println("Stopping controller")
}

func (c *Controller) worker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	ctx := context.Background()

	item, shutdown := c.Queue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.Queue.Done(item)

		// Extract the key from the item in namespace/name format
		key, err := cache.MetaNamespaceKeyFunc(item)
		if err != nil {
			fmt.Println("Error getting key from item:", err)
			// Since we can't process the item, we stop processing it
			c.Queue.Forget(item)
			return nil
		}

		// Split the key into namespace and name
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			fmt.Println("Error splitting key:", err)
			// Since we can't process the item, we stop processing it
			c.Queue.Forget(item)
			return nil
		}

		// Since we only know about the deployment object,
		// We have to check with the API server to determine
		// if the deployment is added or deleted.
		_, err = c.ClientSet.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			err = c.cleanupResources(ctx, ns, name)
			if err != nil {
				// If there is an error, requeue the item
				c.Queue.AddRateLimited(obj)
				fmt.Println("Error cleaning up resources", err)
				return err
			}

			return nil
		} else if err != nil {
			// If there is an error, requeue the item
			c.Queue.AddRateLimited(obj)
			fmt.Println("Error getting deployment info", err)
			return err
		}

		// Expose the deployment
		err = c.exposeDeployment(ctx, ns, name)
		if err != nil {
			// If there is an error, requeue the item
			c.Queue.AddRateLimited(obj)
			fmt.Println("Error exposing deployment", err)
			return err
		}

		c.Queue.Forget(item)

		return nil
	}(item)

	if err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

// cleanupResources will delete the service and ingress
// created by "exposeDeployment".
func (c *Controller) cleanupResources(ctx context.Context, namespace, name string) error {
	err := c.ClientSet.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Service with name %s deleted\n", name)

	err = c.ClientSet.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Ingress with name %s deleted\n", name)

	return nil
}

// exposeDeployment is the main logic of the controller
//
// What it does is whenever a new deployment is added,
// it will create a service to expose the deployment
// and an ingress to route the traffic.
func (c *Controller) exposeDeployment(ctx context.Context, namespace, name string) error {
	// Get the deployment
	deployment, err := c.DeploymentLister.Deployments(namespace).Get(name)
	if err != nil {
		return err
	}

	// Create a service that exposes port 80
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: deployment.Spec.Template.Labels,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
		},
	}
	_, err = c.ClientSet.CoreV1().Services(namespace).Create(ctx, &svc, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Service with name %s created\n", svc.Name)

	// Create an ingress that routes traffic to the service
	ingress := netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{
				{
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: lo.ToPtr(netv1.PathTypePrefix),
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: svc.Name,
											Port: netv1.ServiceBackendPort{
												Name: svc.Spec.Ports[0].Name,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = c.ClientSet.NetworkingV1().Ingresses(namespace).Create(ctx, &ingress, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Ingress with name %s created\n", ingress.Name)

	return nil
}

// handleAdd will be called every time a new deployment is added
func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("Deployment added")
	c.Queue.AddRateLimited(obj)
}

// handleDelete will be called every time a deployment is deleted
func (c *Controller) handleDelete(obj interface{}) {
	fmt.Println("Deployment deleted")
	c.Queue.AddRateLimited(obj)
}
