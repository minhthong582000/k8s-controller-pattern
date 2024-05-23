package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/minhthong582000/k8s-controller-pattern/gitops/internal/controller"
	appclient "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned"
	appinformers "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/informers/externalversions"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/signals"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
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

	gitClient := git.NewGitClient("")
	appClientSet, err := appclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	appInformerFactory := appinformers.NewSharedInformerFactory(appClientSet, time.Second*30)
	stopCh := signals.SetupSignalHandler()
	ctrl := controller.NewController(
		clientSet,
		appClientSet,
		appInformerFactory.Thongdepzai().V1alpha1().Applications(),
		gitClient,
	)
	appInformerFactory.Start(stopCh)
	if err = ctrl.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
		return
	}
}
