#!/bin/bash
kubectl delete service manager
kubectl delete deployment manager-deployment
docker rmi mana -f
docker rmi johnson684/mana:python -f

kubectl delete ksvc image-scale
docker rmi image-scale -f
docker rmi johnson684/image-scale:python -f

kubectl delete ksvc image-recognition
docker rmi image-recognition -f
docker rmi johnson684/image-recognition:python -f

cd manager
docker build -t johnson684/mana:python .
docker push johnson684/mana:python
cd ..

cd image-scale
docker build -t image-scale .
docker tag image-scale:latest johnson684/image-scale:python
docker push johnson684/image-scale:python
cd ..

cd image-recognition
docker build -t image-recognition .
docker tag image-recognition:latest johnson684/image-recognition:python
docker push johnson684/image-recognition:python
cd ..

kubectl apply -f yamls/manager.yaml
kubectl apply -f yamls/apps.yaml
