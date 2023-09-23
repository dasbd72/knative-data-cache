# https://hub.docker.com/_/python
FROM python:3.10-slim AS release

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app
COPY image-scale/. ./
COPY pkg/. ./
RUN pip install -r requirements.txt

EXPOSE 8080

ENTRYPOINT [ "python", "image-scale.py" ]