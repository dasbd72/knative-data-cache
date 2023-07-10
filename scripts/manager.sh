# !/bin/bash

cd manager
docker build -t johnson684/mana:python .
docker push johnson684/mana:python
cd ..

cd manager-go
docker build -t johnson684/mana:golang-socket .
docker push johnson684/mana:golang-socket
cd ..

# delete old manager
kubectl delete -f yamls/manager.yaml

# deploy to k8s
kubectl apply -f yamls/manager.yaml

kubectl port-forward $(kubectl get pods -l "app=manager" -o jsonpath="{.items[0].metadata.name}") 12345:8080

# manager-go
# /
curl -X GET localhost:12345/
# /create
curl -X POST localhost:12345/create -H 'Content-Type: application/json' -d '{"endpoint":"10.121.240.169:9000", "accessKey":"LbtKL76UbWedONnd", "secretKey":"Bt0Omfh0S3ud5VEQAVR85CwinSULl3Sj", "secure": false}'
# /download
curl -X POST localhost:12345/download -H 'Content-Type: application/json' -d '{"endpoint":"10.121.240.169:9000", "bucket":"images-processing", "object":"images/"}'
# /upload
curl -X POST localhost:12345/upload -H 'Content-Type: application/json' -d '{"endpoint":"10.121.240.169:9000", "bucket":"images-processing", "object":"images-scaled/"}'

# curl -X POST localhost:12345/backup -H 'Content-Type: application/json' -d '{"endpoint":"10.121.240.169:9000", "bucket":"images-processing", "object":"images-scaled/"}'
# curl -X POST localhost:12345/download -H 'Content-Type: application/json' -d '{"endpoint":"10.121.240.169:9000", "bucket":"images-processing", "object":"images-scaled/"}'