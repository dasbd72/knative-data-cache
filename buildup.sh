#!/bin/bash
kubectl delete -f yamls/apps.yaml
kubectl delete -f yamls/manager.yaml
kubectl delete -f yamls/cache-deleter.yaml
docker rmi johnson684/mana:golang-socket -f
docker rmi johnson684/image-scale:python-socket -f
docker rmi johnson684/image-recognition:python-socket -f
docker rmi johnson684/cache-deleter:python -f

cd manager-go
docker build -t johnson684/mana:golang-socket .
docker push johnson684/mana:golang-socket
cd ..

cd cache-deleter
docker build -t johnson684/cache-deleter:python .
docker push johnson684/cache-deleter:python
cd ..

docker build -t johnson684/image-scale:python-socket -f  image-scale/Dockerfile .
docker push johnson684/image-scale:python-socket

docker build -t johnson684/image-recognition:python-socket -f  image-recognition/Dockerfile .
docker push johnson684/image-recognition:python-socket

kubectl apply -f yamls/apps.yaml
kubectl apply -f yamls/manager.yaml
kubectl apply -f yamls/cache-deleter.yaml
