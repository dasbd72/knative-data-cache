#!/bin/bash
docker rmi johnson684/manager:latest -f
cd manager-go
docker build --network=host -t johnson684/manager:latest .
docker push johnson684/manager:latest
cd ..

docker rmi dasbd72/data-serve:latest -f
cd data-serve
docker build --network=host -t dasbd72/data-serve:latest .
docker push dasbd72/data-serve:latest
cd ..

docker rmi johnson684/cache-deleter:latest -f
cd cache-deleter
docker build --network=host -t johnson684/cache-deleter:latest .
docker push johnson684/cache-deleter:latest
cd ..

# image chain
docker rmi johnson684/image-scale:latest -f
docker build --network=host -t johnson684/image-scale:latest -f image-scale/Dockerfile .
docker push johnson684/image-scale:latest

docker rmi johnson684/image-recognition:latest -f
docker build --network=host -t johnson684/image-recognition:latest -f image-recognition/Dockerfile .
docker push johnson684/image-recognition:latest

# video chain
docker rmi johnson684/video-split:latest -f
docker build --network=host -t johnson684/video-split:latest -f video-split/Dockerfile .
docker push johnson684/video-split:latest

docker rmi johnson684/video-transcode:latest -f
docker build --network=host -t johnson684/video-transcode:latest -f video-transcode/Dockerfile .
docker push johnson684/video-transcode:latest

docker rmi johnson684/video-merge:latest -f
docker build --network=host -t johnson684/video-merge:latest -f video-merge/Dockerfile .
docker push johnson684/video-merge:latest

