package main

import (
	"flag"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func main() {
	resource := flag.String("resource", "pods", "resource to get GVR for")
	flag.Parse()

	configFlag := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionFlag := cmdutil.NewMatchVersionFlags(configFlag)

	mapper, err := cmdutil.NewFactory(matchVersionFlag).ToRESTMapper()
	if err != nil {
		fmt.Println(err)
		return
	}

	gvr, err := mapper.ResourceFor(
		schema.GroupVersionResource{
			Resource: *resource,
		},
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("{ Group: %s, Version: %s, Resource: %s }\n", gvr.Group, gvr.Version, gvr.Resource)
}
