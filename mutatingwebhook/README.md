# k8s-injector
Kubernetes自动注入服务,向deployment每个容器中注入sidecar, 比如envoy

### 关于server端tls证书操作步骤如下：
- 生成CA机构秘钥(ca.key) & 证书(ca.crt)
    ```text
    openssl req -nodes -new -x509 -days 365 -keyout ca.key -out ca.crt -subj "/CN=EXAMPLE Admission Controller Webhook CA"
    注: CA机构的CN(common name)填写对应介绍即可，无特别要求
    ```
    
- 生成webhook server端私钥(server.key)
    ```text
    openssl genrsa -out tls.key 2048
    ```
    
- 利用CA机构证书&秘钥 生成webhook server端证书(server.crt)
    ```text
    openssl req -new -key tls.key -subj "/CN=p47-mutate-server.wscn-system.svc" | openssl x509 -days 365 -req -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt
      
    注: 此CN一般写{WEBHOOK_SVC_NSME}.{WEBHOOK_SVC_NAMESPACE}.svc格式
        此处不能随意写, api-server会按照MutatingWebhookConfiguration.clientConfig发出
        POST请求，如下例子：就会是 https://p47-mutate-server.kube-system.svc:443/mutate
        所以此时webhook-server端证书一定要是针对p47-mutate-server.kube-system.svc的,否
        则就会报证书不匹配的错误
    ```
    
### 配置MutatingWebhookConfiguration
- caBundle 字段: apiServer用此内容来验证webhook-server的证书
  ```text
  cat ca.crt | base64 | tr -d '\n'
  注: tr -d '\n' 用来去掉换行符

  例如:
  apiVersion: admissionregistration.k8s.io/v1beta1
  kind: MutatingWebhookConfiguration
  metadata:
    name: p47-mutate-server
  webhooks:
    - name: p47-mutate-server.example.com
      namespaceSelector:
        matchLabels:
          name: example
      failurePolicy: Fail
      clientConfig:
        service:
          name: p47-mutate-server
          namespace: wscn-system
          path: "/mutate"
        caBundle: base64(ca.cert)
      rules:
        - operations: ["CREATE"]
          apiGroups: ["apps"]
          apiVersions: ["v1"]
          resources: ["deployments"]
  ```
  
### 开发描述
- 只有response为200,且body满足AdReview的格式,api-server才会更进一步判断,否则会被一律当做错误处理. 错误的处理结果可由failurePolicy设定,值为 Ignore | Fail。Ingore表示忽略这个错误，
api-server继续向下执行该请求。 Fail则表示拒绝掉该请求。比如创建一个deployment，Ignore则可以继续创建，Fail则直接拒绝创建。


