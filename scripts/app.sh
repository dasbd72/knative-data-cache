# !/bin/bash

cd image-scale
docker build -t johnson684/image-scale:python .
docker push johnson684/image-scale:python
cd ..

cd image-recognition
docker build -t johnson684/image-recognition:python .
docker push johnson684/image-recognition:python
cd ..

# deploy to knative
kubectl apply -f yamls/apps.yaml

# first test
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images", "destination":"images-scaled"}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-scaled"}'

# forced remote
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images", "destination":"images-scaled", "force_remote": true}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-scaled", "force_remote": true}'

# second test
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"benchmark_images", "destination":"benchmark_images-scaled"}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"benchmark_images-scaled"}'

# forced remote
time curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"benchmark_images", "destination":"benchmark_images-scaled", "force_remote": true}'
time curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"benchmark_images-scaled", "force_remote": true}'