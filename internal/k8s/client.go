package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient returns a kubernetes client. In-cluster first; falls back to
// kubeconfig (default loading rules) when not running inside a pod.
func NewClient() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig (KUBECONFIG env, $HOME/.kube/config, etc.)
		loader := clientcmd.NewDefaultClientConfigLoadingRules()
		over := &clientcmd.ConfigOverrides{}
		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, over).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("k8s config: %w", err)
		}
	}
	return kubernetes.NewForConfig(cfg)
}
