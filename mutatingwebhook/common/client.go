package common

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

var Kubeset *kubernetes.Clientset

func init() {

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		panic(err)
	}

	var errr error
	Kubeset, errr = kubernetes.NewForConfig(cfg)

	if errr != nil {
		panic(err)
	}
}
