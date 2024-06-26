package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	isInCluster := false

	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	numWorkers := flag.Int("workers", 2, "Number of workers")
	watchNamespace := flag.String("namespace", "default", "Namespace to watch for deployments")
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
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientSet,
		resyncTimeout,
		informers.WithNamespace(*watchNamespace),
	)
	deploymentInformer := informerFactory.Apps().V1().Deployments()

	// Run the controller
	stopCh := setupSignalHandler()
	controller := NewController(clientSet, deploymentInformer)
	informerFactory.Start(stopCh)
	controller.Run(*numWorkers, stopCh)
}

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func setupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}
