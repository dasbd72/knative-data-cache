# Local Storage Manager

## Deploy to knative and kubernetes

Configure knative settings
```bash=
kubectl edit cm config-features -n knative-serving
```

Add the lines right under data as below
```yaml=
data:
  "kubernetes.podspec-persistent-volume-claim": enabled
  "kubernetes.podspec-persistent-volume-write": enabled
  "kubernetes.podspec-init-containers": enalbed
```

Deploy with the script
```bash=
bash deploy.sh
```

## Requests to application

```bash=
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images", "Destination":"images-scaled"}'
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images-scaled"}'
```
