# nginx-globe-gl
Globe.gl visualization for NGINX access logs

## Requirements
- An Elasticsearch user with a role with the following index permissions: `indices:admin/get`, `indices:data/read/search`

## Running locally
Run the following commands
```bash
CAT <<EOF > .env
ELASTICSEARCH_HOST=<replace-host>
ELASTICSEARCH_USER=<replace-user>
ELASTICSEARCH_PASSWORD=<replace-password>
SELF_URL=http://localhost:9999
ELASTICSEARCH_QUERY={"size":1000,"query":{"match":{"kubernetes.labels.app_kubernetes_io/name":"ingress-nginx"}},"sort":[{"@timestamp":"desc"}]}
EOF

go run .
```
Access http://locahost:9999/ 🚀
