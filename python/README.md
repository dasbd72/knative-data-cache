# Python

## Image Scale

```bash=
cd image-scale
```
Build image
```bash=
docker build -t image-scale:python .
```
Run container
```bash=
docker container run -it --rm -p 9090:9090 --name image-scale-py image-scale:python
docker container run -d --rm -p 9090:9090 --name image-scale-py image-scale:python
```
Request
```bash=
curl -X POST 0.0.0.0:9090 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images"}'
```

## Image Recognition

```bash=
cd image-recognition
```
Build image
```bash=
docker build -t image-recognition:python .
```
Run container
```bash=
docker container run -it --rm -p 9091:9091 --name image-recognition-py image-recognition:python
docker container run -d --rm -p 9091:9091 --name image-recognition-py image-recognition:python
```
Request
```bash=
curl -X POST 0.0.0.0:9091 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images-20230501142955/"}'
```