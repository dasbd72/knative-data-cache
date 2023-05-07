import os
import shutil
import json
import time
from datetime import date, datetime

import argparse
import uuid
from minio import Minio
from PIL import Image
from flask import Flask, request, make_response

app = Flask(__name__)

parser = argparse.ArgumentParser(
    prog='Image Recognition',
    description='Runs resnet on the images',
)
parser.add_argument('-p', '--port', type=int, default=9090)
parser.add_argument('--storage_path', type=str, default=None)
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


def copyImages(src_path, dst_path):
    if os.path.exists(dst_path):
        shutil.rmtree(dst_path)
    shutil.copytree(src_path, dst_path)


def scaleImages(local_path, width, height):
    for filename in os.listdir(local_path):
        filepath = os.path.join(local_path, filename)
        img = Image.open(filepath)
        img = img.resize((width, height))
        img.save(filepath, 'JPEG')
    return


@app.route('/', methods=['POST'])
def imageRecognition():
    code_start_time = time.perf_counter()

    data = request.data.decode("utf-8")
    data = json.loads(data)

    if args.storage_path is not None:
        bucket_name = data['Bucket'].rstrip("/")
        download_path = data['Source'].rstrip("/") + "/"
        upload_path = args.storage_path.rstrip("/") + "/" + data['Destination'].rstrip("/") + "/"
        local_path = './storage/'
    else:
        bucket_name = data['Bucket'].rstrip("/")
        download_path = data['Source'].rstrip("/") + "/"
        upload_path = data['Destination'].rstrip("/") + "/"
        local_path = './storage/'

    # remove exist storage and create
    if os.path.exists(local_path):
        shutil.rmtree(local_path)
    os.makedirs(local_path)

    minio_client = Minio(
        endpoint,
        access_key=access_key,
        secret_key=secret_key,
        secure=False
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
    if args.storage_path is not None:
        copyImages(local_path, upload_path)
    else:
        uploadImages(minio_client, bucket_name, local_path, upload_path)

    upload_end_time = time.perf_counter()
    upload_duration = upload_end_time - upload_start_time

    code_end_time = time.perf_counter()
    code_duration = code_end_time - code_start_time
    print(f"Execution time: {code_duration}")
    print(f"Download time: {download_duration}")
    print(f"Scale time: {scale_duration}")
    print(f"Upload time: {upload_duration}")

    # send response
    response = make_response({})
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "image-scale"
    return response


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=args.port)
