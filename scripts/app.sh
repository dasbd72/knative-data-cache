# !/bin/bash

cd image-scale
docker build -t johnson684/image-scale:python .
docker push johnson684/image-scale:python
cd ..

cd image-recognition
docker build -t johnson684/image-recognition:python .
docker push johnson684/image-recognition:python
cd ..

# delete old deployments
kubectl delete -f yamls/apps.yaml

# deploy to knative
kubectl apply -f yamls/apps.yaml

# images
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images", "destination":"images-scaled"}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-scaled", "short_result": true}'

# forced remote
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images", "destination":"images-scaled", "force_remote": true}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-scaled", "force_remote": true, "short_result": true}'

# images-old
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-old", "destination":"images-old-scaled"}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-old-scaled"}'

# forced remote
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-old", "destination":"images-old-scaled", "force_remote": true}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-old-scaled", "force_remote": true}'