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
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images", "Destination":"images-scaled"}'
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images-scaled"}'

# second test
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"benchmark_images", "Destination":"benchmark_images-scaled"}'
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"benchmark_images-scaled"}'
