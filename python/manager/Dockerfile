# https://hub.docker.com/_/python
FROM python:3.10-slim AS release
ENV PYTHONUNBUFFERED True
WORKDIR /app
COPY . ./
RUN pip install -r requirements.txt
EXPOSE 8080
ENTRYPOINT [ "python", "m.py" ]
