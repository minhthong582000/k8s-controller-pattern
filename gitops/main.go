package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	appclient "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	isInCluster := false

	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	// numWorkers := flag.Int("workers", 2, "Number of workers")
	// watchNamespace := flag.String("namespace", "default", "Namespace to watch for deployments")
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
	clientSet, err := appclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	apps, err := clientSet.ThongdepzaiV1alpha1().Applications("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	if len(apps.Items) == 0 {
		fmt.Println("No applications found")
		return
	}

	for _, app := range apps.Items {
		fmt.Printf("Found application %s\n", app.Name)
	}
}
