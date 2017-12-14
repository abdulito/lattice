package servicebuild

import (
	"fmt"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type Controller struct {
	syncHandler func(svcBuildKey string) error
	enqueue     func(svcBuild *crv1.ServiceBuild)

	latticeClient latticeclientset.Interface

	serviceBuildLister       latticelisters.ServiceBuildLister
	serviceBuildListerSynced cache.InformerSynced

	componentBuildLister       latticelisters.ComponentBuildLister
	componentBuildListerSynced cache.InformerSynced

	// recentComponentBuilds holds a map of namespaces which map to a map of definition
	// hashes which map to the name of a ComponentBuild that was recently created
	// in the namespace. recentComponentBuilds should always hold the Name of the most
	// recently created ComponentBuild for a given definition hash.
	// See createComponentBuilds for more information.
	// FIXME: add some GC on this map so it doesn't grow forever (maybe remove in addComponentBuild)
	recentComponentBuildsLock sync.RWMutex
	recentComponentBuilds     map[string]map[string]string

	queue workqueue.RateLimitingInterface
}

func NewController(
	latticeClient latticeclientset.Interface,
	serviceBuildInformer latticeinformers.ServiceBuildInformer,
	componentBuildInformer latticeinformers.ComponentBuildInformer,
) *Controller {
	sbc := &Controller{
		latticeClient:         latticeClient,
		recentComponentBuilds: make(map[string]map[string]string),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-build"),
	}

	sbc.syncHandler = sbc.syncServiceBuild
	sbc.enqueue = sbc.enqueueServiceBuild

	serviceBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addServiceBuild,
		UpdateFunc: sbc.updateServiceBuild,
		// TODO: for now it is assumed that ServiceBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ServiceBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SvcB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.serviceBuildLister = serviceBuildInformer.Lister()
	sbc.serviceBuildListerSynced = serviceBuildInformer.Informer().HasSynced

	componentBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addComponentBuild,
		UpdateFunc: sbc.updateComponentBuild,
		// TODO: for now it is assumed that ComponentBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document CB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.componentBuildLister = componentBuildInformer.Lister()
	sbc.componentBuildListerSynced = componentBuildInformer.Informer().HasSynced

	return sbc
}

func (sbc *Controller) addServiceBuild(obj interface{}) {
	svcb := obj.(*crv1.ServiceBuild)
	glog.V(4).Infof("Adding ServiceBuild %s", svcb.Name)
	sbc.enqueueServiceBuild(svcb)
}

func (sbc *Controller) updateServiceBuild(old, cur interface{}) {
	oldSvcb := old.(*crv1.ServiceBuild)
	curSvcb := cur.(*crv1.ServiceBuild)
	glog.V(4).Infof("Updating ServiceBuild %s", oldSvcb.Name)
	sbc.enqueueServiceBuild(curSvcb)
}

// addComponentBuild enqueues any ServiceBuilds which may be interested in it when
// a ComponentBuild is added.
func (sbc *Controller) addComponentBuild(obj interface{}) {
	cb := obj.(*crv1.ComponentBuild)

	if cb.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		// FIXME: send error event
		return
	}

	glog.V(4).Infof("ComponentBuild %s added.", cb.Name)
	svcbs, err := sbc.getServiceBuildsForComponentBuild(cb)
	if err != nil {
		// FIXME: send error event?
	}
	for _, svcb := range svcbs {
		sbc.enqueueServiceBuild(svcb)
	}
}

// updateComponentBuild enqueues any ServiceBuilds which may be interested in it when
// a ComponentBuild is updated.
func (sbc *Controller) updateComponentBuild(old, cur interface{}) {
	glog.V(5).Info("Got ComponentBuild update")
	oldCb := old.(*crv1.ComponentBuild)
	curCb := cur.(*crv1.ComponentBuild)
	if curCb.ResourceVersion == oldCb.ResourceVersion {
		// Periodic resync will send update events for all known ComponentBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("ComponentBuild ResourceVersions are the same")
		return
	}

	svcbs, err := sbc.getServiceBuildsForComponentBuild(curCb)
	if err != nil {
		// FIXME: send error event?
	}
	for _, svcb := range svcbs {
		sbc.enqueueServiceBuild(svcb)
	}
}

func (sbc *Controller) getServiceBuildsForComponentBuild(cb *crv1.ComponentBuild) ([]*crv1.ServiceBuild, error) {
	svcbs := []*crv1.ServiceBuild{}

	// Find any ServiceBuilds whose ComponentBuildsInfo mention this ComponentBuild
	// TODO: add a cache mapping ComponentBuild Names to active ServiceBuilds which are waiting on them
	//       ^^^ tricky because the informers will start and trigger (aka this method will be called) prior
	//			 to when we could fill the cache

	svcBuilds, err := sbc.serviceBuildLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, svcb := range svcBuilds {
		for _, cbInfo := range svcb.Spec.Components {
			if cbInfo.Name != nil && *cbInfo.Name == cb.Name {
				svcbs = append(svcbs, svcb)
				break
			}
		}
	}

	return svcbs, nil
}

func (sbc *Controller) enqueueServiceBuild(svcb *crv1.ServiceBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(svcb)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", svcb, err))
		return
	}

	sbc.queue.Add(key)
}

func (sbc *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer sbc.queue.ShutDown()

	glog.Infof("Starting service-build controller")
	defer glog.Infof("Shutting down service-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, sbc.serviceBuildListerSynced, sbc.componentBuildListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced.")

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(sbc.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (sbc *Controller) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for sbc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (sbc *Controller) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := sbc.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer sbc.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := sbc.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		sbc.queue.Forget(key)
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
	sbc.queue.AddRateLimited(key)

	return true
}

// syncServiceBuild will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (sbc *Controller) syncServiceBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing ServiceBuild %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing ServiceBuild %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	svcb, err := sbc.serviceBuildLister.ServiceBuilds(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("ServiceBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	stateInfo, err := sbc.calculateState(svcb)
	if err != nil {
		return err
	}

	glog.V(5).Infof("ServiceBuild %v state: %v", svcb.Name, stateInfo.state)

	svcbCopy := svcb.DeepCopy()

	err = sbc.syncComponentBuildStates(svcbCopy, stateInfo)
	if err != nil {
		return nil
	}

	switch stateInfo.state {
	case svcBuildStateHasFailedCBuilds:
		return sbc.syncFailedServiceBuild(svcbCopy, stateInfo.failedCbs)
	case svcBuildStateHasOnlyRunningOrSucceededCBuilds:
		return sbc.syncRunningServiceBuild(svcbCopy, stateInfo.activeCbs)
	case svcBuildStateNoFailuresNeedsNewCBuilds:
		return sbc.syncMissingComponentBuildsServiceBuild(svcbCopy, stateInfo.needsNewCb)
	case svcBuildStateAllCBuildsSucceeded:
		return sbc.syncSucceededServiceBuild(svcbCopy)
	default:
		panic("unreachable")
	}
}
