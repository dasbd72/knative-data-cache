# Python

## Request pipeline

Follow the request commands bellow

## Test Manager

Build image

```bash=
docker build -t dasbd72/test-manager:python ./test-manager
```

Run container

```bash=
docker container run -it --rm -p 8080:8080 --name test-manager dasbd72/test-manager:python
```

## Manager

Build image

```bash=
docker build -t dasbd72/manager:python ./manager
docker push dasbd72/manager:python
```

## Image Scale

Build image

```bash=
docker build -t dasbd72/image-scale:python ./image-scale
docker push dasbd72/image-scale:python
```

Pull image

```bash=
docker pull dasbd72/image-scale:python
```

Run container

```bash=
# Remote storage
docker container run -it --rm -p 9090:8080 --name image-scale-py dasbd72/image-scale:python

# Local storage
docker container run -it --rm -p 9090:8080 -v /dev/shm:/shm --name image-scale-py \
  -e MANAGER_URL="http://$(docker container inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' test-manager):8080" \
  -e STORAGE_PATH="/shm" \
  dasbd72/image-scale:python
```

Request

```bash=
curl -X POST 0.0.0.0:9090 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images", "Destination":"images-scaled"}'
```

## Image Recognition

Build image

```bash=
docker build -t dasbd72/image-recognition:python ./image-recognition
docker push dasbd72/image-recognition:python
```

Pull image

```bash=
docker pull dasbd72/image-recognition:python
```

Run container

```bash=
# Remote storage
docker container run -it --rm -p 9091:8080 --name image-recognition-py dasbd72/image-recognition:python

# Local storage
docker container run -it --rm -p 9091:8080 -v /dev/shm:/shm --name image-recognition-py \
  -e MANAGER_URL="http://$(docker container inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' test-manager):8080" \
  -e STORAGE_PATH="/shm" \
  dasbd72/image-recognition:python
```

Request

```bash=
curl -X POST 0.0.0.0:9091 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images-scaled"}'
```

## Deploy to knative

Configure knative settings

```bash=
kubectl edit cm config-features -n knative-serving
```

Add right below data

```yaml=
data:
  "kubernetes.podspec-persistent-volume-claim": enabled
  "kubernetes.podspec-persistent-volume-write": enabled
  "kubernetes.podspec-init-containers": enalbed
```

```bash=
bash start_service.sh
```

```bash=
curl -X POST http://manager.default.127.0.0.1.sslip.io/download -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Object":"images/"}'
curl -X POST http://image-scale.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images", "Destination":"images-scaled"}'
curl -X POST http://image-recognition.default.127.0.0.1.sslip.io -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images-scaled"}'
```
