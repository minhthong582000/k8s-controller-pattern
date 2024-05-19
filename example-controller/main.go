package main

import (
	"flag"
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	isInCluster := false

	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.Parse()

	// Set up the kubernetes client
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

	// Create an informer
	resyncTimeout := 30 * time.Second

	// Create a shared informer factory that watches resources in all namespaces
	// informerFactory := informers.NewSharedInformerFactory(clientSet, resyncTimeout)

	// Create a shared informer factory that watches resources in the "default" namespace
	namespace := "default"
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientSet,
		resyncTimeout,
		informers.WithNamespace(namespace),
	)
	deploymentInformer := informerFactory.Apps().V1().Deployments()

	// Run the controller
	stopCh := make(chan struct{})
	controller := NewController(clientSet, deploymentInformer)
	informerFactory.Start(stopCh)
	controller.Run(stopCh)
}
