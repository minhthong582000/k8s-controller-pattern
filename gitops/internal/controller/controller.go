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
	k8sutil "github.com/minhthong582000/k8s-controller-pattern/gitops/utils/kube"
	log "github.com/sirupsen/logrus"
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
)

type Controller struct {
	clientSet kubernetes.Interface

	appClientSet appclientset.Interface

	appLister applisters.ApplicationLister

	// Notifies the controller when the cache is synced
	appCacheSync cache.InformerSynced

	// Every time a new event detected by informer, it will be added to the queue
	queue workqueue.RateLimitingInterface

	// appRefreshQueue is used to reconcile the application periodically after created
	appRefreshQueue workqueue.RateLimitingInterface

	gitUtil git.GitClient

	k8sUtil k8sutil.K8s

	eventRecorder record.EventRecorder
}

func NewController(
	clientSet kubernetes.Interface,
	appClientSet appclientset.Interface,
	informer appinformers.ApplicationInformer,
	gitUtil git.GitClient,
	k8sUtil k8sutil.K8s,
) *Controller {
	log.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(log.Debugf)
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
		appRefreshQueue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"application-reconcile",
		),
		gitUtil:       gitUtil,
		k8sUtil:       k8sUtil,
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
	log.Info("Starting controller")

	defer func() {
		log.Debugf("Shutting down controller")
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

	for i := 0; i < numWorkers; i++ {
		// Wait every 1 second to process the next item in the queue
		go wait.Until(c.applicationRefreshWorker, 1*time.Second, stopCh)
	}

	<-stopCh

	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {
	}
}

func (c *Controller) applicationRefreshWorker() {
	for c.processNextAppRefreshItem() {
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

		// Hand it over to the refresh queue on creation
		// TODO: use a cache instead of depending on Informer
		c.requestAppRefresh(app.GetName(), app.GetNamespace())
		c.queue.Forget(obj)

		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		c.updateAppStatus(ctx, obj.(*v1alpha1.Application), &v1alpha1.ApplicationStatus{
			HealthStatus: v1alpha1.HealthStatusDegraded,
		})
	}

	return true
}

func (c *Controller) processNextAppRefreshItem() bool {
	ctx := context.Background()

	appKey, shutdown := c.appRefreshQueue.Get()
	if shutdown {
		return false
	}

	log.Info("Processing application refresh " + appKey.(string))

	// We wrap this block in a func so we can defer c.workqueue.Done.
	app, err := func(appKey string) (*v1alpha1.Application, error) {
		defer c.appRefreshQueue.Done(appKey)

		// Split the key into namespace and name
		ns, name, err := cache.SplitMetaNamespaceKey(appKey)
		if err != nil {
			// Since we can't process the item, we stop processing it
			c.appRefreshQueue.Forget(appKey)
			return nil, fmt.Errorf("error splitting key: %s", err)
		}

		app, err := c.appClientSet.ThongdepzaiV1alpha1().Applications(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			// This means the application is deleted while processing
			if apierrors.IsNotFound(err) {
				c.appRefreshQueue.Forget(appKey)
				return app, nil
			}

			// If there is another type of error, requeue the item
			c.appRefreshQueue.AddRateLimited(appKey)
			return nil, fmt.Errorf("error getting deployment info: %s", err)
		}

		err = c.createResources(ctx, app)
		if err != nil {
			c.appRefreshQueue.AddRateLimited(appKey)
			return nil, fmt.Errorf("error creating resources: %s", err)
		}
		c.appRefreshQueue.Forget(appKey)

		return app, nil
	}(appKey.(string))

	if err != nil {
		utilruntime.HandleError(err)
		c.updateAppStatus(ctx, app, &v1alpha1.ApplicationStatus{
			HealthStatus: v1alpha1.HealthStatusDegraded,
		})
	}

	return true
}

func (c *Controller) createResources(ctx context.Context, app *v1alpha1.Application) error {
	repoPath := path.Join(os.TempDir(), app.Name, strings.Replace(app.Spec.Repository, "/", "_", -1))
	log.Println(repoPath)
	err := c.updateAppStatus(
		ctx,
		app,
		&v1alpha1.ApplicationStatus{
			HealthStatus: v1alpha1.HealthStatusProgressing,
		},
	)
	if err != nil {
		return fmt.Errorf("error updating application status to Processing: %s", err)
	}

	log.WithField("application", app.Name).Info("Creating resources")

	// Clone the repository
	log.Debugf("Cloning repository to %s", repoPath)
	err = c.gitUtil.CloneOrFetch(app.Spec.Repository, repoPath)
	if err != nil {
		return fmt.Errorf("error cloning repository: %s", err)
	}
	log.Debugf("Repository cloned to %s", repoPath)
	sha, err := c.gitUtil.Checkout(repoPath, app.Spec.Revision)
	if err != nil {
		return fmt.Errorf("error checking out revision: %s", err)
	}
	log.Debugf("Checked out revision %s", app.Spec.Revision)

	// Generate manifests
	log.Infof("Generating manifests for application %s", app.Name)
	generatedResources, err := c.k8sUtil.GenerateManifests(path.Join(repoPath, app.Spec.Path))
	if err != nil {
		return fmt.Errorf("error generating manifests: %s", err)
	}

	// Get current resources
	log.Infof("Getting resources for application %s", app.Name)
	label := map[string]string{
		common.LabelKeyAppInstance: app.Name,
	}
	currentResources, err := c.k8sUtil.GetResourceWithLabel(label)
	if err != nil {
		return fmt.Errorf("error getting resources with label: %s, %s", label, err)
	}

	// Set the label for the generated resources
	label = map[string]string{
		common.LabelKeyAppInstance: app.Name,
	}
	err = c.k8sUtil.SetLabelsForResources(generatedResources, label)
	if err != nil {
		return fmt.Errorf("error setting labels for resources: %s", err)
	}

	// Calculate diff
	log.Infof("Diffing resources for application %s", app.Name)
	diff, err := c.k8sUtil.DiffResources(currentResources, generatedResources)
	if err != nil {
		return fmt.Errorf("error diffing resources: %s", err)
	}
	if diff {
		// Create resources
		for _, r := range generatedResources {
			// TODO: use namespace from application spec
			err = c.k8sUtil.CreateResource(ctx, r, app.GetNamespace())
			if err != nil {
				return fmt.Errorf("error creating resources: %s", err)
			}
		}
	} else {
		log.WithField("application", app.Name).Info("No changes in resources")
	}

	err = c.updateAppStatus(
		ctx,
		app,
		&v1alpha1.ApplicationStatus{
			HealthStatus: v1alpha1.HealthStatusHealthy,
			Revision:     sha,
			LastSyncAt:   metav1.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("error updating application status to Ready: %s", err)
	}

	c.eventRecorder.Event(app, corev1.EventTypeNormal, common.SuccessSynced, common.MessageResourceSynced)

	log.WithField("application", app.Name).Info("Resources created")

	return nil
}

func (c *Controller) deleteResources(app *v1alpha1.Application) error {
	if app.Name == "" {
		return fmt.Errorf("application name is empty")
	}

	repoPath := path.Join(os.TempDir(), app.Name, strings.Replace(app.Spec.Repository, "/", "_", -1))

	// Get all resources with the label
	label := map[string]string{
		common.LabelKeyAppInstance: app.Name,
	}
	resources, err := c.k8sUtil.GetResourceWithLabel(label)
	if err != nil {
		return fmt.Errorf("error getting resources with label: %s, %s", label, err)
	}

	log.WithField("application", app.Name).Info("Deleting resources")
	for _, r := range resources {
		err := c.k8sUtil.DeleteResource(context.Background(), r, r.GetNamespace())
		if err != nil {
			return fmt.Errorf("error deleting resources: %s", err)
		}
	}

	err = c.gitUtil.CleanUp(repoPath)
	if err != nil {
		return fmt.Errorf("error cleaning up repository: %s", err)
	}
	log.WithField("application", app.Name).Info("Resources deleted")

	return nil
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Debugf("Application added")

	c.queue.AddRateLimited(obj)
}

func (c *Controller) handleDelete(obj interface{}) {
	log.Debugf("Application deleted")

	c.queue.AddRateLimited(obj)
}

func (c *Controller) handleUdate(old, new interface{}) {
	log.Debugf("Application updated")

	oldApp, oldOk := old.(*v1alpha1.Application)
	newApp, newOk := new.(*v1alpha1.Application)
	if !oldOk || !newOk {
		log.Error("Error decoding object, invalid type")
		return
	}

	// If there are changes in fields other than spec, we don't need to reconcile
	if !equality.Semantic.DeepEqual(oldApp.ObjectMeta, newApp.ObjectMeta) || !equality.Semantic.DeepEqual(oldApp.Status, newApp.Status) {
		return
	}

	// Compare old and new spec
	if equality.Semantic.DeepEqual(oldApp.Spec, newApp.Spec) {
		log.Debugf("No changes in application spec: %s", newApp.Name)
	}

	// TODO: use a cache instead of depending on Informer
	c.requestAppRefresh(newApp.GetName(), newApp.GetNamespace())
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

func (c *Controller) requestAppRefresh(appName string, namespace string) {
	key := namespace + "/" + appName
	c.appRefreshQueue.AddRateLimited(key)
}
