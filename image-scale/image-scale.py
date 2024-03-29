import os
import shutil
import json
import time

import uuid
from wrapper import MinioWrapper as Minio
from PIL import Image
from flask import Flask, request, make_response
import requests

app = Flask(__name__)

# Minio
endpoint = os.getenv("MINIO_ENDPOINT")
access_key = os.getenv("MINIO_ACCESS_KEY")
secret_key = os.getenv("MINIO_SECRET_KEY")


def downloadImages(minio_client: Minio, bucket_name, remote_path, local_path, object_list=[], remote_download=False):
    cnt = 0
    for obj in object_list:
        minio_client.fget_object(bucket_name, remote_path + obj, local_path + os.path.basename(obj), remote_download=remote_download)
        cnt += 1
    return cnt


def uploadImages(minio_client: Minio, bucket_name, local_path, remote_path):
    for filename in os.listdir(local_path):
        filepath = os.path.join(local_path, filename)
        minio_client.fput_object(bucket_name, remote_path + filename, filepath)
    return


def scaleImages(local_path, width, height):
    for filename in os.listdir(local_path):
        filepath = os.path.join(local_path, filename)
        img = Image.open(filepath)
        img = img.resize((width, height))
        img.save(filepath, 'JPEG')
    return


def send_trigger_request(bucket_name, source):
    return None

    # TODO: Fix this, Ryan
    url = "http://broker-ingress.knative-eventing.svc.cluster.local/default/default"
    headers = {
        "Ce-Id": str(uuid.uuid4()),
        "Ce-Specversion": "0.3",
        "Ce-Type": "image-read",
        "Ce-Source": "image-scale",
        "Content-Type": "application/json",
    }
    source_data = source + "-scaled"
    data = {
        "bucket": bucket_name,
        "source": source_data,
        "short_result": True
    }
    try:
        response = requests.post(url, headers=headers, json=data)
        response.raise_for_status()  # Raises an exception for 4xx and 5xx status codes
        return response.text
    except requests.exceptions.RequestException as e:
        print(f"An error occurred: {e}")
        return None


@app.route('/', methods=['POST'])
def imageRecognition():
    code_start_time = time.perf_counter()

    data = request.data.decode("utf-8")
    data = json.loads(data)

    bucket_name = data['bucket'].rstrip("/")
    download_path = data['source'].rstrip("/") + "/"
    object_list = data['object_list']
    upload_path = data['destination'].rstrip("/") + "/"
    if 'force_remote' in data:
        force_remote = data['force_remote']
    else:
        force_remote = False
    local_path = f'./storage-{uuid.uuid4()}/'
    # add
    # remove exist storage and create
    if os.path.exists(local_path):
        shutil.rmtree(local_path)
    os.makedirs(local_path)

    minio_client = Minio(
        endpoint,
        access_key=access_key,
        secret_key=secret_key,
        secure=False,
        force_remote=force_remote
    )
    print(f"Connected to {endpoint}")

    download_start_time = time.perf_counter()
    downloadImages(minio_client, bucket_name, download_path, local_path, object_list=object_list, remote_download=True)
    download_end_time = time.perf_counter()
    download_duration = download_end_time - download_start_time

    scale_start_time = time.perf_counter()
    scaleImages(local_path, 1024, 1024)
    scale_end_time = time.perf_counter()
    scale_duration = scale_end_time - scale_start_time

    upload_start_time = time.perf_counter()
    uploadImages(minio_client, bucket_name, local_path, upload_path)
    upload_end_time = time.perf_counter()
    upload_duration = upload_end_time - upload_start_time

    code_end_time = time.perf_counter()
    code_duration = code_end_time - code_start_time

    # send response
    response = make_response(json.dumps({
        "force_remote": force_remote,
        "code_duration": code_duration,
        "download_duration": download_duration,
        "scale_duration": scale_duration,
        "upload_duration": upload_duration,
    }))
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "image-scale"
    return response


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=os.environ.get('PORT', 8080))
