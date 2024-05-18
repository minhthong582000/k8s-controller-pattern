package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func ListNamespaces(ctx context.Context, clientSet *kubernetes.Clientset) {
	namespaces, err := clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, ns := range namespaces.Items {
		fmt.Println(ns.Name)
	}
}

// Informer is a function that demonstrates how to use informer to watch for changes in the cluster
//
// Why wouldn't we use Watch API? Because using Watch API will put a lot of pressure on the API server
// for a large number of resources.
func Informer(clienSet *kubernetes.Clientset) {
	// After resyncTimeout is reached, the informer will sync
	// its cache with the newest state of the resource.
	resyncTimeout := 30 * time.Second

	// SharedInformerFactory will create informers for resources in all namespaces
	informerFactory := informers.NewSharedInformerFactory(clienSet, resyncTimeout)

	// But sometimes we only want to watch specific resources in a specific namespace
	// informerFactory := informers.NewSharedInformerFactoryWithOptions(
	// 	clienSet,
	// 	resyncTimeout,
	// 	informers.WithTweakListOptions(func(options *metav1.ListOptions) {
	// 		options.LabelSelector = "default"
	// 	}),
	// )

	podInformer := informerFactory.Core().V1().Pods()

	// After new state of the resource is detected, the informer will send
	// the events to the handler based on the type.
	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(new interface{}) {
				fmt.Println("Pod added")
			},
			UpdateFunc: func(old, new interface{}) {
				fmt.Println("Pod updated")
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Println("Pod deleted")
			},
		},
	)

	// Initialize the informer
	informerFactory.Start(wait.NeverStop)

	// Once the informer is initialized, wait for first api call to complete
	informerFactory.WaitForCacheSync(wait.NeverStop)

	// We can use the informer to get the pod
	_, err := podInformer.Lister().Pods("default").Get("default")
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	ctx := context.Background()
	isInCluster := false

	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %s\n", err.Error())
		isInCluster = true
	}

	if isInCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	}

	config.Timeout = 120 * time.Second

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// First example, list all namespaces
	ListNamespaces(ctx, clientSet)

	// Second example, informer
	Informer(clientSet)
}
