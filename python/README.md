# Python

## Request pipeline

Follow the request commands bellow

## Image Scale

Build image
```bash=
docker build -t image-scale:python ./image-scale
```
Run container
```bash=
docker container run -it --rm -p 9090:9090 --name image-scale-py image-scale:python
docker container run -d --rm -p 9090:9090 --name image-scale-py image-scale:python
```
Request
```bash=
curl -X POST 0.0.0.0:9090 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images", "Destination":"images-scaled"}'
```

## Image Recognition

Build image
```bash=
docker build -t image-recognition:python ./image-recognition
```
Run container
```bash=
docker container run -it --rm -p 9091:9090 --name image-recognition-py image-recognition:python
docker container run -d --rm -p 9091:9090 --name image-recognition-py image-recognition:python
```
Request
```bash=
curl -X POST 0.0.0.0:9091 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images-scaled"}'
```