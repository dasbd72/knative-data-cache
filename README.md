# Images Processing Benchmark

## Image scale

Execute
```bash=
go run cmd/image-scale/image-scale.go [-dry false] [-port 9090]
```
Build binary
```bash=
go build -o image-scale cmd/image-scale/image-scale.go
```
Build image
```bash=
docker build -t image-scale -f docker/image-scale.dockerfile .
```
Run container
```bash=
docker container run -it --rm -p 9090:9090 --name image-scale image-scale
docker container run -d --rm -p 9090:9090 --name image-scale image-scale
```
Request
```bash=
curl -X GET 0.0.0.0:9090
curl -X POST 0.0.0.0:9090 -H 'Content-Type: application/json' -d '{"Bucket":"images-processing", "Source":"images"}'
```