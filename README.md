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
# Force remote
time curl -X POST http://image-scale.default.192.168.100.0.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"stress-benchmark", "source":"larger_image_0", "destination":"larger_image_scaled", "force_remote":true, "object_list":["DSC08867.JPG", "DSC08868.JPG", "DSC08869.JPG", "DSC08871.JPG", "DSC08872.JPG"]}'
time curl -X POST http://image-recognition.default.192.168.100.0.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"stress-benchmark", "source":"larger_image_scaled", "short_result": true, "force_remote":true, "object_list":["DSC08867.JPG", "DSC08868.JPG", "DSC08869.JPG", "DSC08871.JPG", "DSC08872.JPG"]}'

curl -X POST http://video-split.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"video-processing", "source":"original-video", "destination":"splitted-video", "object_list":["sample.mp4"]}'
curl -X POST http://video-transcode.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"video-processing", "source":"splitted-video", "destination":"transcoded-video", "object_list":["seg1_sample.mp4"]}'
curl -X POST http://video-merge.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"video-processing", "source":"transcoded-video", "destination":"merged-video", "object_list":["seg1_sample.avi", "seg2_sample.avi", "seg3_sample.avi", "seg4_sample.avi", "seg5_sample.avi"]}'

curl -X POST http://word-count-start.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"word-count", "source":"text", "destination":"word-count-dict", "force_remote":false, "object_list":["big.txt"]}'
```
