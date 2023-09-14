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
  "kubernetes.podspec-init-containers": enabled
```

Deploy with the script
```bash=
bash deploy.sh
```

## Requests to application

```bash=
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images", "destination":"images-scaled"}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-scaled", "short_result": true}'

curl -X POST http://video-split.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"video-processing", "source":"original-video", "destination":"splitted-video"}'
curl -X POST http://video-transcode.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"video-processing", "source":"splitted-video/seg1_sample.mp4", "destination":"transcoded-video"}'
curl -X POST http://video-merge.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"video-processing", "source":"transcoded-video", "destination":"merged-video"}'
```
