package business

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/labstack/echo"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SidecarAnnotationKey = "inject-envoy"
	SidecarContainerName = "envoy"
	ConfigMapName        = "envoy-cnf"
	MutateServerNs = "default"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func MutateHandler(ctx echo.Context) error {
	reqBody, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.String(500, err.Error())
		return nil
	}

	reqAdReview := &v1beta1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(reqBody, nil, reqAdReview); err != nil {
		glog.Errorf("Decode admissionReview failed: %v", err)
		ctx.String(500, err.Error())
		return nil
	}

	rspAdReview := mutate(reqAdReview)
	body, err := json.Marshal(rspAdReview)
	if err != nil {
		glog.Errorf("Marshal response failed: %v", err)
		ctx.String(500, err.Error())
		return nil
	}

	if _, errr := ctx.Response().Write(body); errr != nil {
		ctx.String(500, errr.Error())
	}

	return nil
}

func successfulAdmission(uid types.UID, patchs []byte, patchType *v1beta1.PatchType) *v1beta1.AdmissionReview {
	return &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:       uid,
			Allowed:   true,
			Patch:     patchs,
			PatchType: patchType,
		},
	}
}

func failedAdmission(uid types.UID, err error) *v1beta1.AdmissionReview {
	return &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:       uid,
			Allowed:   false,
			Result:    &metav1.Status{
				Message: err.Error(),
			},
		},
	}
}

func mutate(reqAdReview *v1beta1.AdmissionReview) *v1beta1.AdmissionReview {

	req := reqAdReview.Request
	var deployment appsv1.Deployment
	if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
		glog.Errorf("mutate err: %v", err)
		return failedAdmission(req.UID, err)
	}

	if !mutationRequired(deployment) {
		glog.Infof("mutationRequired not required")
		return successfulAdmission(req.UID, nil, nil)
	}

	patchs, err := getPatchs()
	if err != nil {
		return failedAdmission(req.UID, err)
	}

	patchBytes, err := json.Marshal(patchs)

	if err != nil {
		glog.Errorf("createPatch err: %v", err)
		return failedAdmission(req.UID, err)
	}

	glog.Infof("begin patch: %s", string(patchBytes))

	pt := v1beta1.PatchTypeJSONPatch
	return successfulAdmission(req.UID, patchBytes, &pt)
}

func mutationRequired(deploy appsv1.Deployment) bool {

	//没有开启envoy 或者 其值不等于true, 则跳过注入逻辑
	if val, ok := deploy.Annotations[SidecarAnnotationKey]; !ok || val != "true" {
		glog.Infof("%s skip inject action, annotation not match", deploy.Name)
		return false
	}

	//开启注入的Deployment已经有envoy, 则跳过注入逻辑
	for _, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == SidecarContainerName {
			glog.Infof("%s skip inject action, already sidecar mode", deploy.Name)
			return false
		}
	}

	return true
}

func GetInjectContainerObjects(injectorStr string) ([]corev1.Container, error) {
	objects := make([]corev1.Container, 0)

	var container corev1.Container
	err := json.Unmarshal([]byte(injectorStr), &container)
	if err != nil {
		glog.Errorf("Unmarshal sidecarConf failed: %s", err.Error())
		return nil, err
	}

	objects = append(objects, container)
	return objects, nil
}

// 在原有json中patch形式添加sidecar所描述的容器, 分别放在第0个位置
func getPatchs() ([]patchOperation, error) {
	var patchs []patchOperation

	containerStr, err := getContainerStrFromConfigMap(MutateServerNs, ConfigMapName, SidecarContainerName)
	if err != nil {
		return nil, err
	}
	containerObjects, err := GetInjectContainerObjects(containerStr)
	if err != nil {
		return nil, err
	}

	// patch语法可参考: https://tools.ietf.org/html/rfc6902#page-12
	for _, container := range containerObjects {
		patchs = append(patchs, patchOperation{
			Op:    "add",
			Path:  "/spec/template/spec/containers/0",
			Value: container,
		})
	}
	return patchs, nil
}

func getContainerStrFromConfigMap(ns, configMapName, keyName string) (string, error) {
	configMap, err := GetConfigMap(configMapName, ns)
	if err != nil {
		return "", err
	}

	if configMap == nil {
		glog.Infof("ConfigMap  %s is nil", configMapName)
		return "", fmt.Errorf("ConfigMap %s is nil", configMapName)
	}

	val, _ := configMap.Data[keyName]
	if len(val) != 0 {
		return val, nil
	}

	return "", fmt.Errorf("cannot find key %s", keyName)
}
