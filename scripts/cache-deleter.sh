cd cache-deleter
docker build -t johnson684/cache-deleter:python .
docker push johnson684/cache-deleter:python
cd ..

kubectl delete -f yamls/cache-deleter.yaml

kubectl apply -f yamls/cache-deleter.yaml