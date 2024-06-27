package cmd

import (
	"time"

	"github.com/minhthong582000/k8s-controller-pattern/gitops/internal/controller"
	appclient "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned"
	appinformers "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/informers/externalversions"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/signals"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git"
	k8sutil "github.com/minhthong582000/k8s-controller-pattern/gitops/utils/kube"
	logutil "github.com/minhthong582000/k8s-controller-pattern/gitops/utils/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"
)

var (
	kubeconfig string
	numWorkers int
	logLevel   string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the gitops controller",
	Long:  `Run the gitops controller. Can be run locally with kubeconfig provided or in-cluster.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Set up the logger
		err := logutil.SetUpLogrus(logLevel)
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set up the kubernetes client
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Infof("Failed to load kubeconfig, falling back to in-cluster config...")

			// Fallback to in-cluster config
			config, err = rest.InClusterConfig()
			if err != nil {
				return err
			}
		}
		config.Timeout = 120 * time.Second
		config.QPS = 1000
		config.Burst = 1000

		appClientSet, err := appclient.NewForConfig(config)
		if err != nil {
			return err
		}
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return err
		}

		// Set up the git client
		gitUtil := git.NewGitClient("")
		dynClientSet, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}

		// Set up k8s utility
		discoveryClient := clientSet.Discovery()
		k8sutil := k8sutil.NewK8s(discoveryClient, dynClientSet)

		// Set up the controller
		appInformerFactory := appinformers.NewSharedInformerFactory(appClientSet, time.Second*30)
		stopCh := signals.SetupSignalHandler()
		ctrl := controller.NewController(
			clientSet,
			appClientSet,
			appInformerFactory.Thongdepzai().V1alpha1().Applications(),
			gitUtil,
			k8sutil,
		)
		appInformerFactory.Start(stopCh)
		if err = ctrl.Run(numWorkers, stopCh); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "~/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	runCmd.PersistentFlags().IntVarP(&numWorkers, "workers", "w", 2, "Number of workers")
	runCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error, fatal, panic)")
}
