# nginx-globe-gl
Globe.gl visualization for NGINX access logs

## Requirements
- An Elasticsearch user with a role with the following index permissions: `indices:admin/get`, `indices:data/read/search`
- Installed https://dev.maxmind.com/geoip/updating-databases tools

## Running locally
Run the following commands
```bash
cat <<EOF > .env
ELASTICSEARCH_HOST=<replace-host>
ELASTICSEARCH_USER=<replace-user>
ELASTICSEARCH_PASSWORD=<replace-password>
SELF_URL=http://localhost:9999
ELASTICSEARCH_QUERY={"size":1000,"query":{"match":{"kubernetes.labels.app_kubernetes_io/name":"ingress-nginx"}},"sort":[{"@timestamp":"desc"}]}
EOF

cat <<EOF > GeoIP.conf
AccountID <maxmind-acc-id>
LicenseKey <maxmind-license-key>
EditionIDs GeoLite2-City
EOF

geoipupdate -vf GeoIP.conf -d .

go run .
```
Access http://locahost:9999/ 🚀
