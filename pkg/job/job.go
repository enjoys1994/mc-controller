package job

import (
	"context"
	"errors"
	"k8s.io/client-go/rest"
	"wangguoyan/mc-operator/pkg/cluster"
	"wangguoyan/mc-operator/pkg/controller"
	"wangguoyan/mc-operator/pkg/manager"
)

type WatchJob struct {
	resources  []*WatchResource
	restClient *rest.RESTClient
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

// StartResourceWatch 启动集群监听，如果已经存在对应监听则忽略
func (w *WatchJob) StartResourceWatch(clusters ...ClusterInfoInterface) context.CancelFunc {
	return w.doResourceWatch(clusters...)
}

// 创建并启动指定集群监听
func (w *WatchJob) doResourceWatch(clusterInfos ...ClusterInfoInterface) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	// 因为mgr start方法为阻塞方法，所以需要启动协程执行
	mgr := manager.New()
	var clusters []*cluster.Cluster
	for i := range clusterInfos {
		c := cluster.New(clusterInfos[i].GetClusterName(), GetCfgByClusterInfo(clusterInfos[i]), cluster.Options{})
		clusters = append(clusters, c)
	}
	go func() {
		// 遍历需要监听的列表
		for i := range w.resources {
			resource := w.resources[i]
			co := controller.New(resource.Reconciler, controller.Options{})
			for j := range clusters {
				c := clusters[j]
				if resource.Scheme != nil {
					n := &cluster.Cluster{}
					*n = *c
					n.SetScheme(resource.Scheme)
					c = n
				}
				if err := co.WatchResourceReconcileObject(ctx, c, resource.ObjectType, resource.WatchOptions); err != nil {
					errChan <- err
					return
				}
			}
			mgr.AddController(co)
		}
		err := mgr.Start(ctx)
		errChan <- err
	}()
	return cancel
}
