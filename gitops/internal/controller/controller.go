package controller

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/minhthong582000/k8s-controller-pattern/gitops/common"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/apis/application/v1alpha1"
	appclientset "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned/scheme"
	appinformers "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/informers/externalversions/application/v1alpha1"
	applisters "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/listers/application/v1alpha1"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
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

	eventRecorder record.EventRecorder
}

func NewController(
	clientSet kubernetes.Interface,
	appClientSet appclientset.Interface,
	informer appinformers.ApplicationInformer,
	gitClient git.GitClient,
) *Controller {
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(4).Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: common.ControllerName})

	c := &Controller{
		clientSet:    clientSet,
		appClientSet: appClientSet,
		appLister:    informer.Lister(),
		appCacheSync: informer.Informer().HasSynced,
		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"application",
		),
		gitClient:     gitClient,
		eventRecorder: recorder,
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

func (c *Controller) Run(numWorkers int, stopCh <-chan struct{}) error {
	klog.Info("Starting controller")

	defer func() {
		c.queue.ShutDown()
	}()

	// Wait for the caches to be synced before starting workers
	if !cache.WaitForCacheSync(stopCh, c.appCacheSync) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	for i := 0; i < numWorkers; i++ {
		// Wait every 1 second to process the next item in the queue
		go wait.Until(c.worker, 1*time.Second, stopCh)
	}

	<-stopCh

	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	ctx := context.Background()

	obj, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.queue.Done(obj)

		// Extract the key from the item in namespace/name format
		key, err := cache.MetaNamespaceKeyFunc(obj)
		if err != nil {
			// Since we can't process the item, we stop processing it
			c.queue.Forget(obj)
			return fmt.Errorf("error getting key from item: %s", err)
		}

		// Split the key into namespace and name
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			// Since we can't process the item, we stop processing it
			c.queue.Forget(obj)
			return fmt.Errorf("error splitting key: %s", err)
		}

		// Since we only know about the deployment object,
		// We have to check with the API server to determine
		// if the deployment is added or deleted.
		app, err := c.appClientSet.ThongdepzaiV1alpha1().Applications(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				app = obj.(*v1alpha1.Application)
				err = c.deleteResources(app)
				if err != nil {
					c.queue.AddRateLimited(obj)
					return fmt.Errorf("error cleaning up resources: %s", err)
				}

				c.queue.Forget(obj)
				return nil
			}

			// If there is another type of error, requeue the item
			c.queue.AddRateLimited(obj)
			return fmt.Errorf("error getting deployment info: %s", err)
		}

		err = c.createResources(ctx, app)
		if err != nil {
			c.queue.AddRateLimited(obj)
			return fmt.Errorf("error creating resources: %s", err)
		}

		c.queue.Forget(obj)

		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		c.updateAppStatus(ctx, obj.(*v1alpha1.Application), &v1alpha1.ApplicationStatus{
			Status: "Failed",
		})
	}

	return true
}

func (c *Controller) createResources(ctx context.Context, app *v1alpha1.Application) error {
	repoPath := path.Join(os.TempDir(), app.Name, strings.Replace(app.Spec.Repository, "/", "_", -1))

	err := c.updateAppStatus(
		ctx,
		app,
		&v1alpha1.ApplicationStatus{
			Status: "Processing",
		},
	)
	if err != nil {
		return fmt.Errorf("error updating application status to Processing: %s", err)
	}

	klog.Infof("Creating resources for application %s", app.Name)

	// Clone the repository
	klog.Infof("Cloning repository to %s", repoPath)
	err = c.gitClient.CloneOrFetch(app.Spec.Repository, repoPath)
	if err != nil {
		return fmt.Errorf("error cloning repository: %s", err)
	}
	klog.Infof("Repository cloned to %s", repoPath)
	sha, err := c.gitClient.Checkout(repoPath, app.Spec.Revision)
	if err != nil {
		return fmt.Errorf("error checking out revision: %s", err)
	}
	klog.Infof("Checked out revision %s", app.Spec.Revision)

	// Generate manifests

	err = c.updateAppStatus(
		ctx,
		app,
		&v1alpha1.ApplicationStatus{
			Status:   "Ready",
			Revision: sha,
		},
	)
	if err != nil {
		return fmt.Errorf("error updating application status to Ready: %s", err)
	}

	c.eventRecorder.Event(app, corev1.EventTypeNormal, common.SuccessSynced, common.MessageResourceSynced)

	return nil
}

func (c *Controller) deleteResources(app *v1alpha1.Application) error {
	if app.Name == "" {
		return fmt.Errorf("application name is empty")
	}

	repoPath := path.Join(os.TempDir(), app.Name, strings.Replace(app.Spec.Repository, "/", "_", -1))

	klog.Infof("Deleting resources for application %s", app.Name)
	err := c.gitClient.CleanUp(repoPath)
	if err != nil {
		return fmt.Errorf("error cleaning up repository: %s", err)
	}
	klog.Info("Resources cleaned up")

	return nil
}

func (c *Controller) handleAdd(obj interface{}) {
	klog.Info("Application added")

	c.queue.AddRateLimited(obj)
}

func (c *Controller) handleDelete(obj interface{}) {
	klog.Info("Application deleted")

	c.queue.AddRateLimited(obj)
}

func (c *Controller) handleUdate(old, new interface{}) {
	klog.Info("Application updated")

	oldApp, oldOk := old.(*v1alpha1.Application)
	newApp, newOk := new.(*v1alpha1.Application)
	if !oldOk || !newOk {
		klog.Error("Error decoding object, invalid type")
		return
	}

	// Compare old and new spec
	if equality.Semantic.DeepEqual(oldApp.Spec, newApp.Spec) {
		klog.Info("No changes in spec, skipping")
		return
	}

	c.queue.AddRateLimited(new)
}

// updateAppStatus to safely update the status of an application.
// We need this instead of using UpdateStatus() since the obj can
// be updated between the time we get and do the status modification.
func (c *Controller) updateAppStatus(ctx context.Context, app *v1alpha1.Application, status *v1alpha1.ApplicationStatus) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		queryApp, err := c.appClientSet.ThongdepzaiV1alpha1().Applications(app.Namespace).Get(ctx, app.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		queryApp.Status = *status
		_, err = c.appClientSet.ThongdepzaiV1alpha1().Applications(queryApp.Namespace).UpdateStatus(ctx, queryApp, metav1.UpdateOptions{})
		if err == nil {
			return nil
		}

		return err
	})
	if err != nil {
		return err
	}

	return nil
}
