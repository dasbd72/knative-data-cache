import os
import shutil
import json
import time
from datetime import date, datetime

import argparse
import uuid
from wrapper import MinioWrapper as Minio
from PIL import Image
from flask import Flask, request, make_response
import requests

app = Flask(__name__)

parser = argparse.ArgumentParser(
    prog='Image Scale',
    description='Scales the images to 224x224',
)
parser.add_argument('-p', '--port', type=int, default=8080)
args = parser.parse_args()

endpoint = "10.121.240.169:9000"
access_key = "LbtKL76UbWedONnd"
secret_key = "Bt0Omfh0S3ud5VEQAVR85CwinSULl3Sj"


def downloadImages(minio_client: Minio, bucket_name, remote_path, local_path):
    obj_lst = minio_client.list_objects(bucket_name, remote_path, False)
    cnt = 0
    for obj in obj_lst:
        minio_client.fget_object(bucket_name, obj.object_name, local_path + os.path.basename(obj.object_name))
        cnt += 1
    obj_lst.close()
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
        "bucket":bucket_name, 
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
    source = data['source'].rstrip("/")
    upload_path = data['destination'].rstrip("/") + "/"
    if 'force_remote' in data:
        force_remote = data['force_remote']
    else:
        force_remote = False
    local_path = './storage/'
    # add
    if 'force_backup' in data:
        force_backup = data['force_backup']
    else:
        force_backup = False
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
        force_remote=force_remote,
        force_backup=force_backup
    )
    print(f"Connected to {endpoint}")

    download_start_time = time.perf_counter()
    downloadImages(minio_client, bucket_name, download_path, local_path)
    download_end_time = time.perf_counter()
    download_duration = download_end_time - download_start_time

    scale_start_time = time.perf_counter()
    scaleImages(local_path, 224, 224)
    scale_end_time = time.perf_counter()
    scale_duration = scale_end_time - scale_start_time

    upload_start_time = time.perf_counter()
    uploadImages(minio_client, bucket_name, local_path, upload_path)
    upload_end_time = time.perf_counter()
    upload_duration = upload_end_time - upload_start_time

    minio_client.close()

    code_end_time = time.perf_counter()
    code_duration = code_end_time - code_start_time

    response_text = send_trigger_request(bucket_name, source)
    # send response
    response = make_response(json.dumps({
        "force_remote": force_remote,
        "force_backup": force_backup,
        "code_duration": code_duration,
        "download_duration": download_duration,
        "scale_duration": scale_duration,
        "upload_duration": upload_duration,
        "download_post_duration": minio_client.get_download_perf(),
        "upload_post_duration": minio_client.get_upload_perf(),
        "backup_post_duration": minio_client.get_backup_perf()
    }))
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "image-scale"
    return response


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=args.port)
