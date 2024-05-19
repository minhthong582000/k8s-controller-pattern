package main

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	applisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	ClientSet           kubernetes.Interface
	DeploymentLister    applisters.DeploymentLister
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

func (c *Controller) Run(ch <-chan struct{}) {
	fmt.Println("Starting controller")

	// Wait for the caches to be synced before starting workers
	if !cache.WaitForCacheSync(ch, c.DeploymentCacheSync) {
		fmt.Println("Error syncing cache")
	}

	// Wait every 1 second to process the next item in the queue
	go wait.Until(c.worker, 1*time.Second, ch)

	// Block the main thread to prevent the program from exiting
	<-ch
}

func (c *Controller) worker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	item, shutdown := c.Queue.Get()
	if shutdown {
		return false
	}
	// If everything work as expected, we can forget the item
	defer c.Queue.Forget(item)

	// Extract the key from the item in namespace/name format
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Println("Error getting key from item")
		return false
	}

	// Split the key into namespace and name
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Println("Error splitting key")
		return false
	}

	err = c.exposeDeployment(ns, name)
	if err != nil {
		// TODO: Implement retry logic
		fmt.Println("Error exposing deployment", err)
		return false
	}

	return true
}

// exposeDeployment is the main logic of the controller
//
// What it does is whenever a new deployment is added,
// it will create a service to expose the deployment
// and an ingress to route the traffic.
func (c *Controller) exposeDeployment(namespace, name string) error {
	ctx := context.Background()

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
	fmt.Println("Ingress with Name: ", ingress.Name, " created")

	return nil
}

// handleAdd will be called every time a new deployment is added
func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("Deployment added")
	c.Queue.Add(obj)
}

// handleDelete will be called every time a deployment is deleted
func (c *Controller) handleDelete(obj interface{}) {
	fmt.Println("Deployment deleted")
	c.Queue.Add(obj)
}
