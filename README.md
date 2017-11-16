# k8s-zk-announser

a service to write k8s service nodeport/loadbalancer to zookeeper as (twitter finagle) service set

currently supporting loadbalancer

the service looks for two service annotations
* `service.announser/zookeeper-path` (the full path to were in zookeeper to add a member)
* `service.announser/portname`  (the service ports name that service is running on. limited to 1)

## example setup

the example will result in one nginx service running with a internal elb. The service
will be announsed in zookeeper at path `/aurora/jobs/role/prod/service` all members will 
be added in to that path. as `/aurora/jobs/role/prod/service/member_<sequense>` and
the port will be the service http port taken from `service.announser/portname`
```yaml
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx
        servicePort: "80"
    spec:
      containers:
        - name: web
          image: nginx:1.7.9
          ports:
            - containerPort: 80

---
apiVersion: v1
kind: Service
metadata:
  name: nginx
  annotations:
    service.announser/zookeeper-path: "/aurora/jobs/role/prod/service"
    service.announser/portname: http
    service.beta.kubernetes.io/aws-load-balancer-internal: 0.0.0.0/0
spec:
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: 80
  selector:
    app: nginx
  type: LoadBalancer
```
