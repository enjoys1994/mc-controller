package job

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"wangguoyan/mc-operator/pkg/controller"
	"wangguoyan/mc-operator/pkg/reconcile"
)

// WatchResource 监听资源，包括类型和监听方法
type WatchResource struct {
	ObjectType   client.Object
	Scheme       *runtime.Scheme
	Reconciler   reconcile.Reconciler
	WatchOptions controller.WatchOptions
	Owner        *Owner
}
type Owner struct {
	ObjectType   client.Object
	WatchOptions controller.WatchOptions
}

type ClusterInfoInterface interface {
	GetToken() string
	GetApiServer() string
	RestConfig() *rest.Config
	GetClusterName() string
}

func NewClusterDefault(key string) ClusterInfoInterface {
	return &ClusterInfo{
		key: key,
		cfg: ctl.GetConfigOrDie(),
	}
}
func NewClusterWithCfg(key string, cfg *rest.Config) ClusterInfoInterface {
	return &ClusterInfo{
		key: key,
		cfg: cfg,
	}
}

type ClusterInfo struct {
	token     string
	apiServer string
	cfg       *rest.Config
	key       string
}

func (c *ClusterInfo) GetToken() string {
	return c.token
}
func (c *ClusterInfo) GetApiServer() string {
	return c.apiServer
}
func (c *ClusterInfo) RestConfig() *rest.Config {
	return c.cfg
}
func (c *ClusterInfo) GetClusterName() string {
	return c.key
}
