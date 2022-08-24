package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"log"
	ctl "sigs.k8s.io/controller-runtime"
	"time"
	"wangguoyan/mc-operator/pkg/job"
	"wangguoyan/mc-operator/pkg/reconcile"
)

func main() {
	watchResources := []*job.WatchResource{
		{
			ObjectType: &corev1.Pod{},
			Reconciler: &testReconciler{},
			//Scheme: APi.Scheme, 自定义crd
		},
		{
			ObjectType: &corev1.Pod{},
			Reconciler: &testReconciler{},
			Scheme:     scheme.Scheme,
		},
	}
	watchJob, err := job.NewWatchJob(watchResources)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	watchJob.AddFailedRollBack(func(clusterName string, err error) {
		klog.Infof("cluster %s watch error : %s", clusterName, err.Error())
	})
	// 监听指定集群
	go func() {
		time.Sleep(15 * time.Second)
		// 停止监听指定集群
		watchJob.StopResourceWatch(job.NewClusterDefault("test"))
		go func() {
			time.Sleep(15 * time.Second)
			watchJob.StopResourceWatch(job.NewClusterDefault("test2"))
		}()
	}()
	watchJob.StartResourceWatch(job.NewClusterDefault("test"), job.NewClusterWithCfg("test2", ctl.GetConfigOrDie()))
}

type testReconciler struct {
}

func (r *testReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {

	pod := &corev1.Pod{}
	err := req.GetClient().Get(context.TODO(), types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, pod)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}
	log.Printf("%s / %s /%s /%s", req.Cluster.GetClusterName(), pod.GetName(), pod.GetNamespace(), pod.UID)
	return reconcile.Result{}, nil
}
