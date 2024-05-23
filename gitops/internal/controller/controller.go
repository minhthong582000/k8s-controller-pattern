package controller

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/apis/application/v1alpha1"
	appclientset "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned"
	appinformers "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/informers/externalversions/application/v1alpha1"
	applisters "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/listers/application/v1alpha1"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	clientSet kubernetes.Interface

	appClientSet appclientset.Interface

	appLister applisters.ApplicationLister

	// Notifies the controller when the cache is synced
	appCacheSync cache.InformerSynced

	// Every time a new event detected by informer, it will be added to the queue
	queue workqueue.RateLimitingInterface

	gitClient git.GitClient
}

func NewController(
	clientSet kubernetes.Interface,
	appClientSet appclientset.Interface,
	informer appinformers.ApplicationInformer,
	gitClient git.GitClient,
) *Controller {
	c := &Controller{
		clientSet:    clientSet,
		appClientSet: appClientSet,
		appLister:    informer.Lister(),
		appCacheSync: informer.Informer().HasSynced,
		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"application",
		),
		gitClient: gitClient,
	}

	informer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			UpdateFunc: c.handleUdate,
			DeleteFunc: c.handleDelete,
		},
	)

	return c
}

func (c *Controller) Run(numWorkers int, stopCh <-chan struct{}) {
	fmt.Println("Starting controller")

	defer func() {
		c.queue.ShutDown()
	}()

	// Wait for the caches to be synced before starting workers
	if !cache.WaitForCacheSync(stopCh, c.appCacheSync) {
		fmt.Println("Error syncing cache")
	}

	for i := 0; i < numWorkers; i++ {
		// Wait every 1 second to process the next item in the queue
		go wait.Until(c.worker, 1*time.Second, stopCh)
	}

	<-stopCh
}

func (c *Controller) worker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	ctx := context.Background()

	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		app := obj.(*v1alpha1.Application)
		app = app.DeepCopy()

		defer c.queue.Done(item)

		// Extract the key from the item in namespace/name format
		key, err := cache.MetaNamespaceKeyFunc(item)
		if err != nil {
			// Since we can't process the item, we stop processing it
			c.queue.Forget(item)
			return fmt.Errorf("error getting key from item: %s", err)
		}

		// Split the key into namespace and name
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			// Since we can't process the item, we stop processing it
			c.queue.Forget(item)
			return fmt.Errorf("error splitting key: %s", err)
		}

		// Since we only know about the deployment object,
		// We have to check with the API server to determine
		// if the deployment is added or deleted.
		_, err = c.appClientSet.ThongdepzaiV1alpha1().Applications(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				err = c.deleteResources(app)
				if err != nil {
					c.queue.AddRateLimited(obj)
					return fmt.Errorf("error cleaning up resources: %s", err)
				}

				c.queue.Forget(item)
				return nil
			}

			// If there is another type of error, requeue the item
			c.queue.AddRateLimited(obj)
			fmt.Println("Error getting deployment info", err)
			return err
		}

		err = c.createResources(app)
		if err != nil {
			c.queue.AddRateLimited(obj)
			return fmt.Errorf("error creating resources: %s", err)
		}

		c.queue.Forget(item)

		return nil
	}(item)

	if err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) createResources(app *v1alpha1.Application) error {
	repoPath := path.Join(os.TempDir(), app.Name, strings.Replace(app.Spec.Repository, "/", "_", -1))

	app.Status.Status = "Processing"
	app.Status.Revision = app.Spec.Revision
	_, err := c.appClientSet.ThongdepzaiV1alpha1().Applications(app.Namespace).Update(context.Background(), app, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating application status to Processing: %s", err)
	}

	fmt.Println("Creating resources for application", app.Name)

	// Clone the repository
	fmt.Println("Cloning repository to", repoPath)
	err = c.gitClient.CloneOrFetch(app.Spec.Repository, repoPath)
	if err != nil {
		return fmt.Errorf("error cloning repository: %s", err)
	}
	fmt.Println("Checking out revision")
	err = c.gitClient.Checkout(repoPath, app.Spec.Revision)
	if err != nil {
		return fmt.Errorf("error checking out revision: %s", err)
	}
	fmt.Println("Repository cloned and revision checked out")

	// Generate manifests
	app, err = c.appClientSet.ThongdepzaiV1alpha1().Applications(app.Namespace).Get(context.Background(), app.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting application: %s", err)
	}
	app.Status.Status = "Ready"
	_, err = c.appClientSet.ThongdepzaiV1alpha1().Applications(app.Namespace).Update(context.Background(), app, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating application status to Ready: %s", err)
	}

	return nil
}

func (c *Controller) deleteResources(app *v1alpha1.Application) error {
	repoPath := path.Join(os.TempDir(), strings.Replace(app.Spec.Repository, "/", "_", -1))

	fmt.Println("Cleaning up files in", repoPath)
	err := c.gitClient.CleanUp(repoPath)
	if err != nil {
		return fmt.Errorf("error cleaning up repository: %s", err)
	}

	return nil
}

func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("New application added")

	// Add the object to the queue
	c.queue.AddRateLimited(obj)
}

func (c *Controller) handleDelete(obj interface{}) {
	fmt.Println("Application deleted")

	// Delete the object from the queue
	c.queue.AddRateLimited(obj)
}

func (c *Controller) handleUdate(old, new interface{}) {
	// fmt.Println("Application updated")

	// newApp, ok := new.(*v1alpha1.Application)
	// if !ok {
	// 	fmt.Println("Error decoding object")
	// 	return
	// }

	// _ = newApp.DeepCopy()
}
