package dnscontroller

import (
	"fmt"
	"os"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Controller struct {
	//Contains the controller specific for updating DNS, Watches Address changes.
	syncEndpointUpdate    func(bKey string) error
	enqueueEndpointUpdate func(endpoint *crv1.Endpoint)

	// R/W of these four variables controller by sharedVarsLock
	cnameList       map[string]crv1.Endpoint
	hostLists       map[string]crv1.Endpoint
	hasRecentlyUpdated bool
	recentlyFlushed map[string]crv1.Endpoint

	sharedVarsLock sync.RWMutex

	latticeClient latticeclientset.Interface

	addressLister       latticelisters.EndpointLister
	addressListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	serverConfigPath string
	hostConfigPath string
}

var(
	updateWaitBeforeFlushTimer = 15
)

func NewController(
	serverConfigPath string,
	hostConfigPath string,
	latticeClient latticeclientset.Interface,
	endpointInformer latticeinformers.EndpointInformer,
) *Controller {

	c := &Controller{
		latticeClient: latticeClient,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system"),
	}

	c.serverConfigPath = serverConfigPath
	c.hostConfigPath = hostConfigPath

	c.syncEndpointUpdate = c.SyncEndpointUpdate
	c.enqueueEndpointUpdate = c.EnqueueEndpointUpdate

	endpointInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addEndpoint,
		UpdateFunc: c.updateEndpoint,
		DeleteFunc: c.deleteEndpoint,
	})
	c.addressLister = endpointInformer.Lister()
	c.addressListerSynced = endpointInformer.Informer().HasSynced

	c.cnameList = make(map[string]crv1.Endpoint)
	c.hostLists = make(map[string]crv1.Endpoint)
	c.recentlyFlushed = make(map[string]crv1.Endpoint)

	return c
}

func (c *Controller) EnqueueEndpointUpdate(endp *crv1.Endpoint) {
	glog.V(5).Infof("enqueueing")
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(endp)

	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", endp, err))
		return
	}

	glog.V(5).Infof("q length before %v %v", c.queue.Len(), key)
	c.queue.Add(key)
	glog.V(5).Infof("q length after %v %v", c.queue.Len(), key)
}

func (c *Controller) addEndpoint(obj interface{}) {
	endpoint := obj.(*crv1.Endpoint)
	glog.V(4).Infof("Endpoint %v/%v added", endpoint.Namespace, endpoint.Name)

	if endpoint.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.

		glog.V(5).Infof("Endpoint %v deletion timestamp not null", endpoint.Name)
		c.deleteEndpoint(endpoint)
		return
	}

	c.enqueueEndpointUpdate(endpoint)
}

func (c *Controller) updateEndpoint(old, cur interface{}) {
	oldEndpoint := old.(*crv1.Endpoint)
	curEndpoint := cur.(*crv1.Endpoint)
	glog.V(5).Info("Got Endpoint %v/%v update", curEndpoint.Namespace, curEndpoint.Name)
	if curEndpoint.ResourceVersion == oldEndpoint.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		glog.V(5).Info("Endpoint %v/%v ResourceVersions are the same", curEndpoint.Namespace, curEndpoint.Name)
		return
	}

	c.enqueueEndpointUpdate(curEndpoint)
}

func (c *Controller) deleteEndpoint(obj interface{}) {
	endpoint, ok := obj.(*crv1.Endpoint)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		endpoint, ok = tombstone.Obj.(*crv1.Endpoint)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not an Endpoint %#v", obj))
			return
		}
	}

	c.enqueueEndpointUpdate(endpoint)
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting local-dns controller")
	defer glog.Infof("Shutting down local-dns controller")

	// wait for your secondary caches to fill before starting your work.
	if !cache.WaitForCacheSync(stopCh, c.addressListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (c *Controller) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (c *Controller) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer c.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := c.syncEndpointUpdate(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		c.queue.Forget(key)
		return true
	}

	// there was a failure so be sure to report it.  This method allows for
	// pluggable error handling which can be used for things like
	// cluster-monitoring
	runtime.HandleError(fmt.Errorf("%v failed with : %v", key, err))

	// since we failed, we should requeue the item to work on later.  This
	// method will add a backoff to avoid hotlooping on particular items
	// (they're probably still not going to work right away) and overall
	// controller protection (everything I've done is broken, this controller
	// needs to calm down or it can starve other useful work) cases.
	c.queue.AddRateLimited(key)

	return true
}

func (c *Controller) SyncEndpointUpdate(key string) error {
	glog.V(1).Infof("Called endpoint update")
	defer func() {
		glog.V(4).Infof("Finished endpoint update")
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		return err
	}

	// Cache lookup
	endpoint, err := c.addressLister.Endpoints(namespace).Get(name)

	// Live lookup
	// endpoint, err = c.latticeClient.LatticeV1().Endpoints(namespace).Get(name, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		glog.V(2).Infof("Endpoint %v has been deleted", key)
		return nil
	}

	if err != nil {
		return err
	}

	glog.V(5).Infof("Endpoint %v state: %v", key, endpoint.Status.State)

	if endpoint.Status.State == crv1.EndpointStateCreated {
		// Created, nothing to do.
		return nil
	}

	// Locks sharedVars for the entire duration. This ensures that the hosts and cnames are updated atomically with checking for
	// cache flushes and prevents missed updates.
	c.sharedVarsLock.Lock()

	if !c.hasRecentlyUpdated {
		glog.V(5).Infof("has not updated recently, will flush all updates in %v seconds", updateWaitBeforeFlushTimer)
		// Safe to write to this boolean as we have the write sharedVarsLock.
		c.hasRecentlyUpdated = true
		go time.AfterFunc(time.Second*time.Duration(updateWaitBeforeFlushTimer), c.FlushRewriteDNS)
	}

	endpointPathURL := endpoint.Spec.Path.ToDomain(true)
	glog.V(5).Infof("URL::::%v", endpointPathURL)

	endpoint = endpoint.DeepCopy()

	_, present := c.recentlyFlushed[key]

	if present {
		glog.V(5).Infof("Endpoint %v already updated. Setting state to created...", key)

		delete(c.recentlyFlushed, endpointPathURL)

		endpoint.Status.State = crv1.EndpointStateCreated
		_, err := c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)

		c.sharedVarsLock.Unlock()

		return err
	}

	if endpoint.Spec.ExternalEndpoint != nil {
		glog.V(5).Infof("Updating endpoint...")
		c.cnameList[key] = *endpoint
	}

	if endpoint.Spec.IP != nil {
		glog.V(5).Infof("Updating ip...")
		c.hostLists[key] = *endpoint
	}

	c.sharedVarsLock.Unlock()

	return nil
}

func (c *Controller) FlushRewriteDNS() {
	err := c.RewriteDnsmasqConfig()

	if err != nil {
		println(err)
	}
}

func CreateEmptyFile(filename string) (*os.File, error) {

	_, err := os.Stat(filename)

	if os.IsExist(err) {
		err := os.Remove(filename)

		if err != nil {
			panic(err)
		}
	}

	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0660)
}

func (c *Controller) RewriteDnsmasqConfig() error {

	glog.V(4).Infof("Rewriting config %v, %v... ", c.hostConfigPath, c.serverConfigPath)

	cname_file, err := CreateEmptyFile(c.serverConfigPath)

	if err != nil {
		return err
	}

	hosts_file, err := CreateEmptyFile(c.hostConfigPath)

	if err != nil {
		return err
	}

	defer func() {
		glog.V(4).Infof("Closing config file...")
		cname_file.Sync()
		hosts_file.Sync()
		cname_file.Close()
		hosts_file.Close()

		// Finished writing to the cache - can now unset the timer flag
		c.hasRecentlyUpdated = false
		c.sharedVarsLock.Unlock()
	}()

	// This is an extra config file, so contains only the options which must be rewritten.
	// Condition on cname is that it exists in the specified host file, or references another cname.
	// Each cname entry of the form cname=ALIAS,...(addn alias),TARGET
	// Full specification here: http://www.thekelleys.org.uk/dnsmasq/docs/dnsmasq-man.html

	// This logic takes a write lock for the entire duration of the update to simplify the logic and to prevent possible missed updates.
	// A missed update is possible when the files are written, and a read lock is released. Before a write lock is acquired to update
	// c.hasRecentlyUpdated, SyncEndpointUpdate takes the write lock and sees that c.hasRecentlyUpdated is false, then proceeds
	// to add an endpoint to the buffer, and releases the lock. Then this function acquires the lock and sets hasRecentlyUpdates to false.
	// If that was the last SyncEndpointUpdate, those updates will never be written to disk as this function cannot be guaranteed
	// to run again.
	c.sharedVarsLock.Lock()

	for _, v := range c.cnameList {
		cname := *v.Spec.ExternalEndpoint
		path := v.Spec.Path.ToDomain(true)
		_, err := cname_file.WriteString("cname=" + path + "," + cname + "\n")
		glog.V(5).Infof("cname=" + path + "," + cname + "\n")

		if err != nil {
			return err
		}
	}

	for _, v := range c.hostLists {
		endpoint := *v.Spec.IP
		path := v.Spec.Path.ToDomain(true)
		_, err := hosts_file.WriteString(endpoint + " " + path + "\n")
		glog.V(5).Infof(endpoint + " " + path + "\n")

		if err != nil {
			return err
		}
	}

	//Now update state and requeue as successful.
	for k, v := range c.cnameList {
		c.recentlyFlushed[k] = v
		c.EnqueueEndpointUpdate(&v)
	}

	for k, v := range c.hostLists {
		c.recentlyFlushed[k] = v
		c.EnqueueEndpointUpdate(&v)
	}

	return nil
}
