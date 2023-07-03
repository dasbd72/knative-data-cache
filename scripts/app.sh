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
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images", "destination":"images-scaled"}' | python -m json.tool
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-scaled", "short_result": true}' | python -m json.tool

# forced remote
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images", "destination":"images-scaled", "force_remote": true}' | python -m json.tool
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-scaled", "force_remote": true, "short_result": true}' | python -m json.tool

# images-old
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-old", "destination":"images-old-scaled"}' | python -m json.tool
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-old-scaled"}' | python -m json.tool

# forced remote
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-old", "destination":"images-old-scaled", "force_remote": true}' | python -m json.tool
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"images-old-scaled", "force_remote": true}' | python -m json.tool

# larger_image
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"larger_image", "destination":"larger_image-scaled"}' | python -m json.tool
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"larger_image-scaled"}' | python -m json.tool

# forced remote
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"larger_image", "destination":"larger_image-scaled", "force_remote": true}' | python -m json.tool
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"bucket":"images-processing", "source":"larger_image-scaled", "force_remote": true}' | python -m json.tool
