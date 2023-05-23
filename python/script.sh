#!/bin/bash
kubectl delete ksvc manager
docker rmi manager -f
docker rmi johnson684/manager:python -f

kubectl delete ksvc image-scale
docker rmi image-scale -f
docker rmi johnson684/image-scale:python -f

kubectl delete ksvc image-recognition
docker rmi image-recognition -f
docker rmi johnson684/image-recognition:python -f

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

cd manager
docker build -t manager .
docker tag manager:latest johnson684/manager:python
docker push johnson684/manager:python
cd ..
