# https://hub.docker.com/_/python
FROM python:3.10-slim AS release

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /python
COPY video-base/. ./
RUN pip install -r requirements.txt
RUN apt-get update && apt-get install -y ffmpeg
