package job

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func CreateClusterClientset(cfg *rest.Config) *kubernetes.Clientset {
	setConfigDefaults(cfg)
	return kubernetes.NewForConfigOrDie(cfg)
}

func CreateClusterRESTClient(cluster ClusterInfoInterface) (*rest.RESTClient, error) {
	cfg := GetCfgByClusterInfo(cluster)
	setConfigDefaults(cfg)
	return rest.RESTClientFor(cfg)
}

func setConfigDefaults(config *rest.Config) {
	config.APIPath = "/apis"
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
}

func GetCfgByClusterInfo(info ClusterInfoInterface) *rest.Config {
	if info.RestConfig() != nil {
		return info.RestConfig()
	}
	tlsClientConfig := rest.TLSClientConfig{
		Insecure: true,
	}
	cfg := rest.Config{
		Host:            info.GetApiServer(),
		TLSClientConfig: tlsClientConfig,
		BearerToken:     info.GetToken(),
	}

	return &cfg
}
