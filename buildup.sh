#!/bin/bash
docker rmi johnson684/manager:latest -f
cd manager-go
docker build -t johnson684/manager:latest .
docker push johnson684/manager:latest
cd ..

docker rmi dasbd72/data-serve:latest -f
cd data-serve
docker build -t dasbd72/data-serve:latest .
docker push dasbd72/data-serve:latest
cd ..

docker rmi johnson684/cache-deleter:latest -f
cd cache-deleter
docker build -t johnson684/cache-deleter:latest .
docker push johnson684/cache-deleter:latest
cd ..

# image chain
docker rmi johnson684/image-scale:latest -f
docker build -t johnson684/image-scale:latest -f image-scale/Dockerfile .
docker push johnson684/image-scale:latest

docker rmi johnson684/image-recognition:latest -f
docker build -t johnson684/image-recognition:latest -f image-recognition/Dockerfile .
docker push johnson684/image-recognition:latest

# video chain
docker rmi johnson684/video-split:latest -f
docker build -t johnson684/video-split:latest -f video-split/Dockerfile .
docker push johnson684/video-split:latest

docker rmi johnson684/video-transcode:latest -f
docker build -t johnson684/video-transcode:latest -f video-transcode/Dockerfile .
docker push johnson684/video-transcode:latest

docker rmi johnson684/video-merge:latest -f
docker build -t johnson684/video-merge:latest -f video-merge/Dockerfile .
docker push johnson684/video-merge:latest

# pv
kubectl apply -f yamls/pv.yaml
kubectl apply -f yamls/pvc.yaml

# delete
# apps
kubectl delete -f yamls/app-image-chain.yaml
kubectl delete -f yamls/app-video-chain.yaml

# controller
kubectl delete -f yamls/manager.yaml
kubectl delete -f yamls/data-serve.yaml
kubectl delete -f yamls/cache-deleter.yaml

# add
# controller
kubectl apply -f yamls/manager.yaml
kubectl apply -f yamls/data-serve.yaml
kubectl apply -f yamls/cache-deleter.yaml

# apps
kubectl apply -f yamls/app-image-chain.yaml
kubectl apply -f yamls/app-video-chain.yaml
