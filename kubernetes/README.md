# Kubernetes
This folder contains basic manifests for deploying nginx-globe-gl in Kubernetes

```bash
kubectl create secret generic nginx-globe-gl \
  --from-literal=SELF_URL= \
  --from-literal=ELASTICSEARCH_HOST= \
  --from-literal=ELASTICSEARCH_USER= \
  --from-literal=ELASTICSEARCH_PASSWORD= \
  --from-literal=ELASTICSEARCH_QUERY= 

kubectl apply -f .

cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-globe-gl
spec:
  rules:
  - host: $SELF_URL
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nginx-globe-gl
            port:
              number: 80
EOF
```
