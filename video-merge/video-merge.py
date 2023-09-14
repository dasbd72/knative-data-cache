import os
import shutil
import json
import time

import uuid
from wrapper import MinioWrapper as Minio
from PIL import Image
from flask import Flask, request, make_response
import requests
from pathlib import Path
import subprocess

app = Flask(__name__)

# Minio
endpoint = os.getenv("MINIO_ENDPOINT")
access_key = os.getenv("MINIO_ACCESS_KEY")
secret_key = os.getenv("MINIO_SECRET_KEY")


def downloadVideo(minio_client: Minio, bucket_name, remote_path, local_path):
    obj_lst = minio_client.list_objects(bucket_name, remote_path, False)
    cnt = 0
    for obj in obj_lst:
        minio_client.fget_object(
            bucket_name, obj.object_name, local_path + os.path.basename(obj.object_name)
        )
        cnt += 1
    obj_lst.close()
    return cnt


def uploadVideo(minio_client: Minio, bucket_name, local_path, remote_path):
    for filename in os.listdir(local_path):
        filepath = os.path.join(local_path, filename)
        minio_client.fput_object(bucket_name, remote_path + filename, filepath)
    return


def mergeVideo(local_path):
    input_files = [f for f in os.listdir(local_path) if f.endswith(".avi")]
    input_files.sort()
    output_path = os.path.join(local_path, "merged.avi")
    f = open("list.txt", "w")
    for file in input_files:
        f.write("file " + os.path.join(local_path, file) + "\n")
    f.close()
    ffmpeg_command = [
        "ffmpeg",
        "-f",
        "concat",
        "-safe",
        "0",
        "-i",
        "list.txt",
        "-c",
        "copy",
        output_path,
    ]
    subprocess.run(ffmpeg_command)
    for file in input_files:
        os.remove(os.path.join(local_path, file))
    os.remove("list.txt")
    return


@app.route("/", methods=["POST"])
def videoMerge():
    code_start_time = time.perf_counter()

    data = request.data.decode("utf-8")
    data = json.loads(data)

    bucket_name = data["bucket"].rstrip("/")
    download_path = data["source"].rstrip("/") + "/"
    upload_path = data["destination"].rstrip("/") + "/"
    if "force_remote" in data:
        force_remote = data["force_remote"]
    else:
        force_remote = False
    local_path = f"./storage-{uuid.uuid4()}/"
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
    )
    print(f"Connected to {endpoint}")

    download_start_time = time.perf_counter()
    downloadVideo(minio_client, bucket_name, download_path, local_path)
    download_end_time = time.perf_counter()
    download_duration = download_end_time - download_start_time

    merge_start_time = time.perf_counter()
    mergeVideo(local_path)
    merge_end_time = time.perf_counter()
    merge_duration = merge_end_time - merge_start_time

    upload_start_time = time.perf_counter()
    uploadVideo(minio_client, bucket_name, local_path, upload_path)
    upload_end_time = time.perf_counter()
    upload_duration = upload_end_time - upload_start_time

    code_end_time = time.perf_counter()
    code_duration = code_end_time - code_start_time

    # send response
    response = make_response(
        json.dumps(
            {
                "force_remote": force_remote,
                "code_duration": code_duration,
                "download_duration": download_duration,
                "merge_duration": merge_duration,
                "upload_duration": upload_duration,
            }
        )
    )
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "image-scale"
    return response


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=os.environ.get("PORT", 8080))
