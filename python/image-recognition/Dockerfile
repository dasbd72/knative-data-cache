# https://hub.docker.com/_/python
FROM python:3.10-slim AS release

# Allow statements and log messages to immediately appear in the Knative logs
ENV PYTHONUNBUFFERED True

WORKDIR /app
COPY . ./
RUN pip install -r requirements.txt
RUN pip3 install torch torchvision torchaudio --extra-index-url https://download.pytorch.org/whl/cpu

EXPOSE 8080

ENTRYPOINT [ "python" ]
CMD [ "image-recognition.py" ]