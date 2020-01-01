package business

import (
	"github.com/golang/glog"
	"gitlab.wallstcn.com/infrastructure/k8s-injector/common"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetConfigMap(name, namespace string) (*v1.ConfigMap, error) {
	client := common.Kubeset
	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("GetConfigMap failed: %v", err)
		return nil, err
	}

	return configMap, nil
}
