# Python

## Request pipeline

Follow the request commands bellow

## Image Scale

Build image
```bash=
docker build -t dasbd72/image-scale:python ./image-scale
```
Pull image
```bash=
docker pull dasbd72/image-scale:python
```
Run container
```bash=
# Remote storage
docker container run -it --rm -p 9090:8080 --name image-scale-py dasbd72/image-scale:python
docker container run -d --rm -p 9090:8080 --name image-scale-py dasbd72/image-scale:python

# Local storage
docker container run -v /dev/shm:/shm -it --rm -p 9090:8080 --name image-scale-py dasbd72/image-scale:python --storage_path /shm
docker container run -v /dev/shm:/shm -d  --rm -p 9090:8080 --name image-scale-py dasbd72/image-scale:python --storage_path /shm

docker container run -v /home/jerry2022/tmp:/disk -it --rm -p 9090:8080 --name image-scale-py dasbd72/image-scale:python --storage_path /disk
docker container run -v /home/jerry2022/tmp:/disk -d  --rm -p 9090:8080 --name image-scale-py dasbd72/image-scale:python --storage_path /disk
```
Request
```bash=
curl -X POST 0.0.0.0:9090 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images", "Destination":"images-scaled"}'
```

## Image Recognition

Build image
```bash=
docker build -t dasbd72/image-recognition:python ./image-recognition
```
Pull image
```bash=
docker pull dasbd72/image-recognition:python
```
Run container
```bash=
# Remote storage
docker container run -it --rm -p 9091:8080 --name image-recognition-py dasbd72/image-recognition:python
docker container run -d --rm -p 9091:8080 --name image-recognition-py dasbd72/image-recognition:python

# Local storage
docker container run -v /dev/shm:/shm -it --rm -p 9091:8080 --name image-recognition-py dasbd72/image-recognition:python --storage_path /shm
docker container run -v /dev/shm:/shm -d  --rm -p 9091:8080 --name image-recognition-py dasbd72/image-recognition:python --storage_path /shm

docker container run -v /home/jerry2022/tmp:/disk -it --rm -p 9091:8080 --name image-recognition-py dasbd72/image-recognition:python --storage_path /disk
docker container run -v /home/jerry2022/tmp:/disk -d  --rm -p 9091:8080 --name image-recognition-py dasbd72/image-recognition:python --storage_path /disk
```
Request
```bash=
curl -X POST 0.0.0.0:9091 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images-scaled"}'
```