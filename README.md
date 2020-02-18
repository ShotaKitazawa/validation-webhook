# validation-webhook

Prohibit updating of immutable fields (`Ingress.metadata.annotations["kubernetes.io/ingress.global-static-ip-name"]`).

## Environment

* Kubernetes v1.15.4
    * `ValidatingWebhookConfiguration` apiVersion: `admissionregistration.k8s.io/v1beta1`

## Demo

* Deploy ValidatingWebhookConfiguration & more

```
$ kubectl apply -f webhook.yaml
eployment.apps/immutable-checker created
service/immutable-checker created
secret/immutable-checker.default.svc created
validatingwebhookconfiguration.admissionregistration.k8s.io/immutable-checker created
```

* Deploy demo Ingress ver1

```
$ kubectl apply -f demo/ingress-01.yaml
ingress.extensions/demo-ingress created
```

* Deploy demo Ingress ver2

```
$ kubectl apply -f demo/ingress-02.yaml
Error from server: error when applying patch:
{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"extensions/v1beta1\",\"kind\":\"Ingress\",\"metadata\":{\"annotations\":{\"kubernetes.io/ingress.global-static-ip-name\":\"demo-gip-02\"},\"name\":\"demo-ingress\",\"namespace\":\"default\"},\"spec\":{\"backend\":{\"serviceName\":\"demo-svc\",\"servicePort\":80},\"rules\":[{\"host\":\"demo.local\",\"http\":{\"paths\":[{\"backend\":{\"serviceName\":\"demo-svc\",\"servicePort\":80}}]}}]}}\n","kubernetes.io/ingress.global-static-ip-name":"demo-gip-02"}}}
to:
Resource: "extensions/v1beta1, Resource=ingresses", GroupVersionKind: "extensions/v1beta1, Kind=Ingress"
Name: "demo-ingress", Namespace: "default"
Object: &{map["apiVersion":"extensions/v1beta1" "kind":"Ingress" "metadata":map["annotations":map["kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"extensions/v1beta1\",\"kind\":\"Ingress\",\"metadata\":{\"annotations\":{\"kubernetes.io/ingress.global-static-ip-name\":\"demo-gip-01\"},\"name\":\"demo-ingress\",\"namespace\":\"default\"},\"spec\":{\"backend\":{\"serviceName\":\"demo-svc\",\"servicePort\":80},\"rules\":[{\"host\":\"demo.local\",\"http\":{\"paths\":[{\"backend\":{\"serviceName\":\"demo-svc\",\"servicePort\":80}}]}}]}}\n" "kubernetes.io/ingress.global-static-ip-name":"demo-gip-01"] "creationTimestamp":"2020-02-18T08:40:31Z" "generation":'\x01' "name":"demo-ingress" "namespace":"default" "resourceVersion":"3425790" "selfLink":"/apis/extensions/v1beta1/namespaces/default/ingresses/demo-ingress" "uid":"82a32170-f5b5-4234-ba4a-9870e46c2402"] "spec":map["backend":map["serviceName":"demo-svc" "servicePort":'P'] "rules":[map["host":"demo.local" "http":map["paths":[map["backend":map["serviceName":"demo-svc" "servicePort":'P']]]]]]] "status":map["loadBalancer":map[]]]}
for: "demo/ingress-02.yaml": admission webhook "immutable-checker.default.svc" denied the request: Ingress.metadata.annotations['kubernetes.io/ingress.global-static-ip-name']: this field is immutable
```

## AdmissionWebhook's Detail

* https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/

