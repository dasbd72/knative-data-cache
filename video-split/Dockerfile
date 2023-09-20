# https://hub.docker.com/_/python
FROM dasbd72/video-base:latest AS release

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app
COPY video-split/. ./
COPY pkg/wrapper.py ./
RUN pip install -r requirements.txt
EXPOSE 8080

ENTRYPOINT [ "python", "video-split.py" ]