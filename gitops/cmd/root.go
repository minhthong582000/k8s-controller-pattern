package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitops",
	Short: "GitOps controller for Kubernetes",
	Long:  `GitOps controller for Kubernetes. This controller watches for changes in a Git repository and applies them to a Kubernetes cluster.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
