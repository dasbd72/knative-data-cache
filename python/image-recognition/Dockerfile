# https://hub.docker.com/_/python
FROM dasbd72/pytorch-cpu AS release

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app
COPY . ./
RUN pip install -r requirements.txt

EXPOSE 8080

ENTRYPOINT [ "python", "image-recognition.py" ]