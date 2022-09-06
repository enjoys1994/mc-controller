# mc-controller(multicluster-controller)
A Go library for building Kubernetes controllers that need to watch resources in multiple clusters.


## Installation
 ```
 $ go get github.com/wangguoyan/mc-controller@v1.0.1
 
  ```


## Usage 

  ```

func main() {
	watchResources := []*job.WatchResource{
		{
			ObjectType: &v1.Deployment{},
			Reconciler: &testReconciler{},
			//Scheme: APi.Scheme, 自定义crd
			Owner: &job.Owner{
				ObjectType:   &v1.ReplicaSet{},
				WatchOptions: controller.WatchOptions{},
			},
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

	go func() {
		time.Sleep(15 * time.Second)
		// 停止监听指定集群
		watchJob.StopResourceWatch(job.NewClusterDefault("test"))
		go func() {
			time.Sleep(15 * time.Second)
			watchJob.StopResourceWatch(job.NewClusterDefault("test2"))
		}()
	}()
	// 开始监听指定集群
	watchJob.StartResourceWatch(job.NewClusterDefault("test"), job.NewClusterDefault("test2"))
}

type testReconciler struct {
}

func (r *testReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {

	obj := &v1.Deployment{}
	err := req.GetClient().Get(context.TODO(), types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	log.Printf("%s / %s /%s /%s", req.Cluster.GetClusterName(), obj.GetName(), obj.GetNamespace(), obj.UID)
	return reconcile.Result{}, nil
}

  ```
