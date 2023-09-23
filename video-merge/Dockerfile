# https://hub.docker.com/_/python
FROM dasbd72/video-base:latest AS release

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app
COPY video-merge/. ./
COPY pkg/. ./
EXPOSE 8080

ENTRYPOINT [ "python", "video-merge.py" ]