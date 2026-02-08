package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient creates a new Kubernetes client using the default kubeconfig rules.
// It follows the same discovery order as kubectl:
// 1. $KUBECONFIG environment variable
// 2. ~/.kube/config
// If cluster is non-empty, it overrides the cluster in the current-context
// (equivalent to kubectl --cluster flag).
func NewClient(cluster string) (*kubernetes.Clientset, *rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	configOverrides := &clientcmd.ConfigOverrides{}
	if cluster != "" {
		configOverrides.Context.Cluster = cluster
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	// RESTConfig取得
	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	// clientset作成
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}

	return clientset, restConfig, nil
}
