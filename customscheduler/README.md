## 原理
- 当创建Pod时, 若Pod.spec.schedulerName未设置, 则此时就采用默认的调度器即kube-scheduler. 若我们设定了该值,则kube-scheduler就会跳过该Pod调度
  自定义的scheduler在监听到该Pod的Add事件后，判断spec.schedulerName是否为自身，如果为自身则调用kubernetes client的bind端口，将该Pod绑定至指定
  Node节点上，从而完成自定义的调度。
- 需要注意, 如果Pod.spec.schedulerName设定了一个错误值，那该Pod就不会被成功调度了。  

## 部署
- 将custom-scheduler部署为一个普通的deployment即可。
- 本地调试开发设定KUBE_CONFIG_PATH为kubeconfig文件路径即可, 集群中部署设定为空即可

## 说明
- 本地可用