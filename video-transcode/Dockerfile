# https://hub.docker.com/_/python
FROM python:3.10-slim AS release

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app
COPY video-transcode/. ./
COPY pkg/wrapper.py ./
RUN pip install -r requirements.txt
RUN apt-get update && apt-get install -y ffmpeg
EXPOSE 8080

ENTRYPOINT [ "python", "video-transcode.py" ]