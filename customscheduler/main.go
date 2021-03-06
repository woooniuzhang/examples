package main

import (
	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"time"
)

const (
	CUSTOM_SCHEDULER_NAME = "example"
	MASTER_URL = ""
	KUBE_CONFIG_PATH = "/Users/woniuxiaoan/.kube/config"
)

var client kubernetes.Interface

type PodEventHandler struct {}
func(pe *PodEventHandler) OnAdd(obj interface{}) {
	newPod, ok := obj.(*v1.Pod)
	if !ok {
		glog.Errorf("get obj failed")
		return
	}

	if _, ok := newPod.Labels["woniu"]; ok {
		glog.Info("--- add handler begin ---")
		glog.Info(*newPod)
		glog.Info("--- add handler end ---")
	}
	//errBuilder := strings.Builder{}
	//newPod, ok := obj.(*v1.Pod)
	//if ok && newPod.Spec.SchedulerName == CUSTOM_SCHEDULER_NAME && len(newPod.Spec.NodeName) == 0 {
	//	glog.Infof("Begin binding pod: %s", newPod.Name)
	//
	//	//绑定操作是一个很宽泛的定义，比如将Pod绑定在某一个Node上，可以利用这个实现自己的scheduler
	//	//又比如将PV绑定在某个一PVC上、某一个ServiceAccount绑定在某个Role上
	//	if err := client.CoreV1().Pods(newPod.Namespace).Bind(&v1.Binding{
	//		ObjectMeta: metav1.ObjectMeta{
	//			Namespace: newPod.Namespace,
	//			Name: newPod.Name,
	//			UID: newPod.UID,
	//		},
	//		Target: v1.ObjectReference{
	//			Kind: "Node",
	//			Name: "10.2.0.8",
	//		},
	//	}); err != nil {
	//		glog.Errorf("Binding pod %s to node %s failed: %v", newPod.Name, "10.2.0.8", err)
	//		return
	//	}
	//}
	//
	//if !ok {
	//	errBuilder.WriteString("convert to pod failed \n")
	//}
	//
	//if newPod.Spec.SchedulerName != CUSTOM_SCHEDULER_NAME {
	//	errBuilder.WriteString(fmt.Sprintf("pod %s scheduler is %s\n", newPod.Name, newPod.Spec.SchedulerName))
	//}
	//
	//if len(newPod.Spec.NodeName) != 0 {
	//	errBuilder.WriteString("pod nodename is not null\n")
	//}
	//
	//glog.Infof("skip bind action, because: %s", errBuilder.String())
}
func(pe *PodEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		glog.Errorf("get obj failed")
		return
	}

	if _, ok := newPod.Labels["woniu"]; ok {
		glog.Info("--- update handler begin ---")
		glog.Info("oldPod: ", *oldPod)
		glog.Info("newPod: ", *newPod)
		glog.Info("--- update handler end ---")
	}
}
func(pe *PodEventHandler) OnDelete(obj interface{}){
	pod, ok := obj.(*v1.Pod)
	if !ok {
		glog.Errorf("get obj failed")
		return
	}

	if _, ok := pod.Labels["woniu"]; ok {
		glog.Info("--- delete handler begin ---")
		glog.Info(*pod)
		glog.Info("--- delete handler end ---")
	}
}

//初始化api-server客户端
//api-server根据etcd自身的watch接口实现了自己的watch接口，避免了各组件直接接触etcd
func init() {
	cfg, err := clientcmd.BuildConfigFromFlags(MASTER_URL, KUBE_CONFIG_PATH)
	if err != nil {
		glog.Fatalf("Build kubernetes client failed: %v", err)
	}
	client = kubernetes.NewForConfigOrDie(cfg)
}

func main() {

	stopCh := signals.SetupSignalHandler()
	sharedInformer := informers.NewSharedInformerFactory(client, time.Second * 10)

	//这个步骤会将Pod注册进shareInformer.informers中
	podInformer := sharedInformer.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(new(PodEventHandler))


	//为会每一个注册的informer启动一个goroutine用于同步数据,until stopCh receives some signal
	sharedInformer.Start(stopCh)
	select{}
}