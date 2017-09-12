package servicebuild

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

const componentBuildDefinitionHashMetadataKey = "lattice-component-build-definition-hash"

type ServiceBuildController struct {
	provider string

	syncHandler func(bKey string) error
	enqueue     func(cb *crv1.ServiceBuild)

	latticeResourceRestClient rest.Interface
	kubeClient                clientset.Interface

	configStore       cache.Store
	configStoreSynced cache.InformerSynced
	configSetChan     chan struct{}
	configSet         bool
	configLock        sync.RWMutex
	config            crv1.ComponentBuildConfig

	serviceBuildStore       cache.Store
	serviceBuildStoreSynced cache.InformerSynced

	componentBuildStore       cache.Store
	componentBuildStoreSynced cache.InformerSynced

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

func NewServiceBuildController(
	provider string,
	kubeClient clientset.Interface,
	latticeResourceRestClient rest.Interface,
	configInformer cache.SharedInformer,
	serviceBuildInformer cache.SharedInformer,
	componentBuildInformer cache.SharedInformer,
) *ServiceBuildController {
	sbc := &ServiceBuildController{
		provider:                  provider,
		latticeResourceRestClient: latticeResourceRestClient,
		kubeClient:                kubeClient,
		recentComponentBuilds:     make(map[string]map[string]string),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "component-build"),
	}

	sbc.syncHandler = sbc.syncServiceBuild
	sbc.enqueue = sbc.enqueueServiceBuild

	configInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    sbc.addConfig,
		UpdateFunc: sbc.updateConfig,
		// TODO: for now we'll assume that Config is never deleted
	})
	sbc.configStore = configInformer.GetStore()
	sbc.configStoreSynced = configInformer.HasSynced

	sbc.configSetChan = make(chan struct{})

	serviceBuildInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addServiceBuild,
		UpdateFunc: sbc.updateServiceBuild,
		// TODO: for now it is assumed that ServiceBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.serviceBuildStore = serviceBuildInformer.GetStore()
	sbc.serviceBuildStoreSynced = serviceBuildInformer.HasSynced

	componentBuildInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addComponentBuild,
		UpdateFunc: sbc.updateComponentBuild,
		// TODO: for now it is assumed that ComponentBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document CB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.componentBuildStore = componentBuildInformer.GetStore()
	sbc.componentBuildStoreSynced = componentBuildInformer.HasSynced

	return sbc
}

func (sbc *ServiceBuildController) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer sbc.queue.ShutDown()

	glog.Infof("Starting component-build controller")
	defer glog.Infof("Shutting down component-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, sbc.configStoreSynced, sbc.componentBuildStoreSynced, sbc.componentBuildStoreSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	// wait for config to be set
	<-sbc.configSetChan

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

func (sbc *ServiceBuildController) addConfig(obj interface{}) {
	config := obj.(*crv1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	sbc.configLock.Lock()
	defer sbc.configLock.Unlock()
	sbc.config = config.DeepCopy().Spec.ComponentBuild

	if !sbc.configSet {
		sbc.configSet = true
		close(sbc.configSetChan)
	}
}

func (sbc *ServiceBuildController) updateConfig(old, cur interface{}) {
	oldConfig := old.(*crv1.Config)
	curConfig := cur.(*crv1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	sbc.configLock.Lock()
	defer sbc.configLock.Unlock()
	sbc.config = curConfig.DeepCopy().Spec.ComponentBuild
}

func (sbc *ServiceBuildController) addServiceBuild(obj interface{}) {
	sBuild := obj.(*crv1.ServiceBuild)
	glog.V(4).Infof("Adding ServiceBuild %s", sBuild.Name)
	sbc.enqueueServiceBuild(sBuild)
}

func (sbc *ServiceBuildController) updateServiceBuild(old, cur interface{}) {
	oldSBuild := old.(*crv1.ServiceBuild)
	curSBuild := cur.(*crv1.ServiceBuild)
	glog.V(4).Infof("Updating ComponentBuild %s", oldSBuild.Name)
	sbc.enqueueServiceBuild(curSBuild)
}

// addComponentBuild enqueues any ServiceBuilds which may be interested in it when
// a ComponentBuild is added.
func (sbc *ServiceBuildController) addComponentBuild(obj interface{}) {
	cBuild := obj.(*crv1.ComponentBuild)

	if cBuild.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		// FIXME: send error event
		return
	}

	glog.V(4).Infof("ComponentBuild %s added.", cBuild.Name)
	for _, sBuild := range sbc.getServiceBuildsForComponentBuild(cBuild) {
		sbc.enqueueServiceBuild(sBuild)
	}
}

// updateComponentBuild enqueues any ServiceBuilds which may be interested in it when
// a ComponentBuild is updated.
func (sbc *ServiceBuildController) updateComponentBuild(old, cur interface{}) {
	glog.V(5).Info("Got ComponentBuild update")
	oldCBuild := old.(*crv1.ComponentBuild)
	curCBuild := cur.(*crv1.ComponentBuild)
	if curCBuild.ResourceVersion == oldCBuild.ResourceVersion {
		// Periodic resync will send update events for all known ComponentBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("ComponentBuild ResourceVersions are the same")
		return
	}

	for _, sBuild := range sbc.getServiceBuildsForComponentBuild(curCBuild) {
		sbc.enqueueServiceBuild(sBuild)
	}
}

func (sbc *ServiceBuildController) getServiceBuildsForComponentBuild(cBuild *crv1.ComponentBuild) []*crv1.ServiceBuild {
	sBuilds := []*crv1.ServiceBuild{}

	// Find any ServiceBuilds whose ComponentBuildInfos mention this ComponentBuild
	// TODO: add a cache mapping ComponentBuild Names to active ServiceBuilds which are waiting on them
	//       ^^^ tricky because the informers will start and trigger (aka this method will be called) prior
	//			 to when we could fill the cache
	for _, sBuildObj := range sbc.serviceBuildStore.List() {
		sBuild := sBuildObj.(*crv1.ServiceBuild)

		for _, cBuildInfo := range sBuild.Spec.ComponentBuildInfos {
			if cBuildInfo.Name != nil && *cBuildInfo.Name == cBuild.Name {
				sBuilds = append(sBuilds, sBuild)
				break
			}
		}
	}

	return sBuilds
}

func (sbc *ServiceBuildController) enqueueServiceBuild(sBuild *crv1.ServiceBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sBuild)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", sBuild, err))
		return
	}

	sbc.queue.Add(key)
}

func (sbc *ServiceBuildController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for sbc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (sbc *ServiceBuildController) processNextWorkItem() bool {
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

// syncServiceBuild will sync the ServiceBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (sbc *ServiceBuildController) syncServiceBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing ServiceBuild %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing ServiceBuild %q (%v)", key, time.Now().Sub(startTime))
	}()

	sBuildObj, exists, err := sbc.serviceBuildStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("ServiceBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	sBuild := sBuildObj.(*crv1.ServiceBuild)

	needsNewCBuildIdx := []int{}
	hasFailedCBuild := false
	hasActiveCBuild := false
	for idx, cBuildInfo := range sBuild.Spec.ComponentBuildInfos {
		exists, err := sbc.componentBuildExists(&cBuildInfo, sBuild.Namespace)
		if err != nil {
			return err
		}

		if !exists {
			needsNewCBuildIdx = append(needsNewCBuildIdx, idx)
			continue
		}

		successful, err := sbc.componentBuildSuccessful(&cBuildInfo, sBuild.Namespace)
		if err != nil {
			return err
		}

		if successful {
			continue
		}

		failed, err := sbc.componentBuildFailed(&cBuildInfo, sBuild.Namespace)
		if err != nil {
			return err
		}

		if failed {
			// No need to do any more work if one of our CBuilds failed.
			hasFailedCBuild = true
			break
		}

		// Component build is pending, queued, or running
		hasActiveCBuild = true
	}

	sBuildCopy := sBuild.DeepCopy()

	// If a ComponentBuild for the ServiceBuild has failed, there's no need to create
	// any more ComponentBuilds, we can simply fail the ServiceBuild.
	// If there are no new ComponentBuilds to create, we can simply make sure the
	// ServiceBuild.Status is up to date.
	if hasFailedCBuild || len(needsNewCBuildIdx) == 0 {
		return sbc.syncServiceBuildStatus(sBuildCopy, hasFailedCBuild, hasActiveCBuild)
	}

	// Create any ComponentBuilds that need to be created and sync ServiceBuild.Status.
	return sbc.createComponentBuilds(sBuildCopy, needsNewCBuildIdx)
}

func (sbc *ServiceBuildController) componentBuildExists(cBuildInfo *crv1.ServiceBuildComponentBuildInfo, namespace string) (bool, error) {
	_, exists, err := sbc.getComponentBuildFromInfo(cBuildInfo, namespace)
	return exists, err
}

func (sbc *ServiceBuildController) componentBuildSuccessful(cBuildInfo *crv1.ServiceBuildComponentBuildInfo, namespace string) (bool, error) {
	cBuild, exists, err := sbc.getComponentBuildFromInfo(cBuildInfo, namespace)
	if err != nil || !exists {
		return false, err
	}

	return cBuild.Status.State == crv1.ComponentBuildStateSucceeded, nil
}

func (sbc *ServiceBuildController) componentBuildFailed(cBuildInfo *crv1.ServiceBuildComponentBuildInfo, namespace string) (bool, error) {
	cBuild, exists, err := sbc.getComponentBuildFromInfo(cBuildInfo, namespace)
	if err != nil || !exists {
		return false, err
	}

	return cBuild.Status.State == crv1.ComponentBuildStateFailed, nil
}

func (sbc *ServiceBuildController) createComponentBuilds(sBuild *crv1.ServiceBuild, needsNewCBuildIdx []int) error {
	for _, newCBuildIdx := range needsNewCBuildIdx {
		cBuildInfo := &sBuild.Spec.ComponentBuildInfos[newCBuildIdx]

		// TODO: is json marshalling of a struct deterministic in order? If not could potentially get
		//		 different SHAs for the same definition. This is OK in the correctness sense, since we'll
		//		 just be duplicating work, but still not ideal
		cBuildDefinitionBlockJson, err := json.Marshal(cBuildInfo.DefinitionBlock)
		if err != nil {
			return err
		}

		h := sha256.New()
		if _, err = h.Write(cBuildDefinitionBlockJson); err != nil {
			return err
		}
		//definitionHash := string(h.Sum(nil))
		definitionHash := hex.EncodeToString(h.Sum(nil))
		cBuildInfo.DefinitionHash = &definitionHash

		cBuild, err := sbc.findComponentBuildForDefinitionHash(sBuild.Namespace, definitionHash)
		if err != nil {
			return err
		}

		// Found an existing ComponentBuild.
		if cBuild != nil && cBuild.Status.State != crv1.ComponentBuildStateFailed {
			cBuildInfo.Name = &cBuild.Name
			continue
		}

		// Existing ComponentBuild failed. We'll try it again.
		var previousCBuildName *string
		if cBuild != nil {
			previousCBuildName = &cBuild.Name
		}

		cBuild, err = sbc.createComponentBuild(sBuild.Namespace, cBuildInfo, previousCBuildName)
		if err != nil {
			return err
		}

		cBuildInfo.Name = &cBuild.Name
	}

	response := &crv1.ServiceBuild{}
	err := sbc.latticeResourceRestClient.Put().
		Namespace(sBuild.Namespace).
		Resource(crv1.ServiceBuildResourcePlural).
		Name(sBuild.Name).
		Body(sBuild).
		Do().
		Into(response)

	if err != nil {
		return err
	}

	// Enqueue the ServiceBuild so we can update its status based on the ComponentBuild statuses.
	sbc.enqueueServiceBuild(response)
	return nil
}

func (sbc *ServiceBuildController) findComponentBuildForDefinitionHash(namespace, definitionHash string) (*crv1.ComponentBuild, error) {
	// Check recent build cache
	cBuild, err := sbc.findComponentBuildInRecentBuildCache(namespace, definitionHash)
	if err != nil {
		return nil, err
	}

	// If we found a build in the recent build cache return it.
	if cBuild != nil {
		return cBuild, nil
	}

	// We couldn't find a ComponentBuild in our recent builds cache, so we'll check the Store to see if
	// there's a ComponentBuild we could use in there.
	return sbc.findComponentBuildInStore(namespace, definitionHash), nil
}

func (sbc *ServiceBuildController) findComponentBuildInRecentBuildCache(namespace, definitionHash string) (*crv1.ComponentBuild, error) {
	sbc.recentComponentBuildsLock.RLock()
	defer sbc.recentComponentBuildsLock.RUnlock()

	cBuildNamespaceCache, ok := sbc.recentComponentBuilds[namespace]
	if !ok {
		return nil, nil
	}

	cBuildName, ok := cBuildNamespaceCache[definitionHash]
	if !ok {
		return nil, nil
	}

	// Check to see if this build is in our ComponentBuild store
	cBuildObj, exists, err := sbc.componentBuildStore.GetByKey(namespace + "/" + cBuildName)
	if err != nil {
		return nil, err
	}

	// The ComponentBuild exists in our store, so return that cached version.
	if exists {
		return cBuildObj.(*crv1.ComponentBuild), nil
	}

	// The ComponentBuild isn't in our store, so we'll need to read from the API
	cBuild := &crv1.ComponentBuild{}
	err = sbc.latticeResourceRestClient.Get().
		Namespace(namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(cBuildName).
		Do().
		Into(cBuild)

	if err != nil {
		if errors.IsNotFound(err) {
			// FIXME: send warn event, this shouldn't happen
			return nil, nil
		}
		return nil, err
	}

	return cBuild, nil
}

func (sbc *ServiceBuildController) findComponentBuildInStore(namespace, definitionHash string) *crv1.ComponentBuild {
	// TODO: similar scalability concerns to getServiceBuildsForComponentBuild
	for _, cBuildObj := range sbc.componentBuildStore.List() {
		cBuild := cBuildObj.(*crv1.ComponentBuild)
		cBuildHashLabel := getComponentBuildDefinitionHashFromLabel(cBuild)
		if cBuildHashLabel == nil {
			// FIXME: add warn event
			continue
		}

		if *cBuildHashLabel == definitionHash && cBuild.Status.State != crv1.ServiceBuildStateFailed {
			return cBuild
		}
	}

	return nil
}

func (sbc *ServiceBuildController) createComponentBuild(
	namespace string,
	cBuildInfo *crv1.ServiceBuildComponentBuildInfo,
	previousCBuildName *string,
) (*crv1.ComponentBuild, error) {
	sbc.recentComponentBuildsLock.Lock()
	defer sbc.recentComponentBuildsLock.Unlock()

	if cBuildNamespaceCache, ok := sbc.recentComponentBuilds[namespace]; ok {
		// If there is a new entry in the recent build cache, a different service build has attempted to
		// build the same component. We'll use this ComponentBuild as ours.
		cBuildName, ok := cBuildNamespaceCache[*cBuildInfo.DefinitionHash]
		if ok && (previousCBuildName == nil || cBuildName != *previousCBuildName) {
			return sbc.getComponentBuildFromApi(namespace, cBuildName)
		}
	}

	// If there is no new entry in the build cache, create a new ComponentBuild.
	cBuild := getNewComponentBuildFromInfo(cBuildInfo)
	result := &crv1.ComponentBuild{}
	err := sbc.latticeResourceRestClient.Post().
		Namespace(namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Body(cBuild).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	cBuildNamespaceCache, ok := sbc.recentComponentBuilds[namespace]
	if !ok {
		cBuildNamespaceCache = map[string]string{}
		sbc.recentComponentBuilds[namespace] = cBuildNamespaceCache
	}
	cBuildNamespaceCache[*cBuildInfo.DefinitionHash] = cBuild.Name

	return result, nil
}

func (sbc *ServiceBuildController) syncServiceBuildStatus(sBuild *crv1.ServiceBuild, hasFailedCBuild, hasActiveCBuild bool) error {
	newStatus := calculateComponentBuildStatus(hasFailedCBuild, hasActiveCBuild)

	if reflect.DeepEqual(sBuild.Status, newStatus) {
		return nil
	}

	sBuild.Status = newStatus

	err := sbc.latticeResourceRestClient.Put().
		Namespace(sBuild.Namespace).
		Resource(crv1.ServiceBuildResourcePlural).
		Name(sBuild.Name).
		Body(sBuild).
		Do().
		Into(nil)

	return err
}

func calculateComponentBuildStatus(hasFailedCBuild, hasActiveCBuild bool) crv1.ServiceBuildStatus {
	if hasFailedCBuild {
		return crv1.ServiceBuildStatus{
			State: crv1.ServiceBuildStateFailed,
		}
	}

	if hasActiveCBuild {
		return crv1.ServiceBuildStatus{
			State: crv1.ServiceBuildStateRunning,
		}
	}

	return crv1.ServiceBuildStatus{
		State: crv1.ServiceBuildStateSucceeded,
	}
}
