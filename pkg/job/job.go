package job

import (
	"context"
	"errors"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sync"
	"wangguoyan/mc-operator/pkg/cluster"
	"wangguoyan/mc-operator/pkg/controller"
	"wangguoyan/mc-operator/pkg/manager"
	"wangguoyan/mc-operator/pkg/util"
)

type WatchJob struct {
	resources   []*WatchResource
	restClient  *rest.RESTClient
	mgrs        util.ThreadSafeMap
	ctx         context.Context
	cancel      context.CancelFunc
	contexts    util.ThreadSafeMap
	cancels     util.ThreadSafeMap
	failedHooks []func(clusterName string, err error)
}

func NewWatchJob(res []*WatchResource) (*WatchJob, error) {
	if len(res) == 0 {
		return nil, errors.New("watch resource is empty")
	}
	watchJob := &WatchJob{
		resources: res,
	}
	return watchJob, nil
}

func (w *WatchJob) AddFailedRollBack(f ...func(clusterName string, err error)) *WatchJob {
	for i := range f {
		w.failedHooks = append(w.failedHooks, f[i])
	}
	return w
}

func (w *WatchJob) StartResourceWatch(clusters ...ClusterInfoInterface) {
	if clusters == nil || len(clusters) == 0 {
		klog.Errorf("cluster should be nil")
	}
	w.doResourceWatch(clusters...)
}

func (w *WatchJob) StopResourceWatch(clusters ...ClusterInfoInterface) {
	for i := range clusters {
		c := clusters[i]
		v, ok := w.cancels.Load(c.GetClusterName())
		if ok {
			v.(context.CancelFunc)()
		}
		w.cancels.Delete(c.GetClusterName())
		w.contexts.Delete(c.GetClusterName())
	}
	if w.cancels.Size() == 0 {
		w.cancel()
	}
}

func (w *WatchJob) StopWatch() {
	w.cancel()
}

func (w *WatchJob) getCtxForClusterName(name string) context.Context {
	v, ok := w.contexts.Load(name)
	if ok {
		return v.(context.Context)
	}
	ctx, cancel := context.WithCancel(w.ctx)
	w.contexts.Store(name, ctx)
	w.cancels.Store(name, cancel)
	return ctx
}
func (w *WatchJob) getMgrByClusterName(name string) *manager.Manager {
	v, ok := w.mgrs.Load(name)
	if ok {
		return v.(*manager.Manager)
	}
	mgr := manager.New()
	w.mgrs.Store(name, mgr)
	return mgr
}

// 创建并启动指定集群监听
func (w *WatchJob) doResourceWatch(clusterInfos ...ClusterInfoInterface) {
	w.ctx, w.cancel = context.WithCancel(context.Background())

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(clusterInfos))
	go func() {
		// 遍历需要监听的列表
		for i := range w.resources {
			resource := w.resources[i]
			for i := range clusterInfos {
				co := controller.New(resource.Reconciler, controller.Options{})
				c := cluster.New(clusterInfos[i].GetClusterName(), GetCfgByClusterInfo(clusterInfos[i]), cluster.Options{})
				if resource.Scheme != nil {
					c.SetScheme(resource.Scheme)
				}
				if err := co.WatchResourceReconcileObject(w.getCtxForClusterName(c.GetClusterName()), c, resource.ObjectType, resource.WatchOptions); err != nil {
					for i := range w.failedHooks {
						w.failedHooks[i](c.GetClusterName(), err)
					}
					waitGroup.Done()
					continue
				}
				w.getMgrByClusterName(c.GetClusterName()).AddController(co)
			}
		}
		w.mgrs.Range(func(key interface{}, value interface{}) (shouldContinue bool) {
			go func() {
				if err := value.(*manager.Manager).Start(w.getCtxForClusterName(key.(string))); err != nil {
					klog.Errorf("start controller failed, err :%s", err.Error())
					for i := range w.failedHooks {
						w.failedHooks[i](key.(string), err)
					}
				}
				waitGroup.Done()
			}()
			return true
		})
	}()

	waitGroup.Wait()
}
