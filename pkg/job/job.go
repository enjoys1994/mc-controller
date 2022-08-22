package job

import (
	"context"
	"errors"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
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

func (w *WatchJob) StartResourceWatch(clusters ...ClusterInfoInterface) context.CancelFunc {
	return w.doResourceWatch(clusters...)
}

// 创建并启动指定集群监听
func (w *WatchJob) doResourceWatch(clusterInfos ...ClusterInfoInterface) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)

	mgr := manager.New()
	go func() {
		// 遍历需要监听的列表
		for i := range w.resources {
			resource := w.resources[i]
			co := controller.New(resource.Reconciler, controller.Options{})
			for i := range clusterInfos {
				c := cluster.New(clusterInfos[i].GetClusterName(), GetCfgByClusterInfo(clusterInfos[i]), cluster.Options{})
				if resource.Scheme != nil {
					c.SetScheme(resource.Scheme)
				}
				if err := co.WatchResourceReconcileObject(ctx, c, resource.ObjectType, resource.WatchOptions); err != nil {
					errChan <- err
					return
				}
			}
			mgr.AddController(co)
		}
		if err := mgr.Start(ctx); err != nil {
			klog.Errorf("start controller failed, err :%s", err.Error())
		}
	}()
	return cancel
}
