# mc-controller(multicluster-controller)
A Go library for building Kubernetes controllers that need to watch resources in multiple clusters.


## Usage 

  ```
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

	// 监听指定集群
	_ = watchJob.StartResourceWatch(job.NewClusterDefault("test"), job.NewClusterDefault("test2"))
	go func() {
		time.Sleep(15 * time.Second)
		// 停止监听指定集群
		watchJob.StopResourceWatch(job.NewClusterDefault("test"))
		go func() {
			time.Sleep(15 * time.Second)
			watchJob.StopResourceWatch(job.NewClusterDefault("test2"))
		}()
	}()
	time.Sleep(10000000 * time.Second)

type testReconciler struct {
}

func (r *testReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {

	pod := &corev1.Pod{}
	err := req.GetClient().Get(context.TODO(), types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, pod)
	if err != nil {
		return reconcile.Result{}, err
	}
	log.Printf("%s / %s /%s /%s", req.Cluster.GetClusterName(), pod.GetName(), pod.GetNamespace(), pod.UID)
	return reconcile.Result{}, nil
}

  ```
   
