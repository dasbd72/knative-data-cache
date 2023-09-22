import os
import shutil
import json
import time

import uuid
from wrapper import MinioWrapper as Minio
from PIL import Image
from flask import Flask, request, make_response
import requests
from moviepy.editor import VideoFileClip
import subprocess

app = Flask(__name__)

# Minio
endpoint = os.getenv("MINIO_ENDPOINT")
access_key = os.getenv("MINIO_ACCESS_KEY")
secret_key = os.getenv("MINIO_SECRET_KEY")


def downloadVideos(minio_client: Minio, bucket_name, remote_path, local_path, object_list=[]):
    cnt = 0
    for obj in object_list:
        minio_client.fget_object(bucket_name, remote_path + obj, local_path + os.path.basename(obj))
        cnt += 1
    return cnt


def uploadVideos(minio_client: Minio, bucket_name, local_path, remote_path):
    for filename in os.listdir(local_path):
        filepath = os.path.join(local_path, filename)
        minio_client.fput_object(bucket_name, remote_path + filename, filepath)
    return


def splitVideo(local_path):
    l = os.listdir(local_path)
    for filename in l:
        filepath = os.path.join(local_path, filename)
        video = VideoFileClip(filepath)
        total_duration = video.duration
        num_segments = 5
        segment_duration = total_duration / num_segments
        for i in range(num_segments):
            start_time = i * segment_duration
            output_file = os.path.join(local_path, f"seg{i+1}_{filename}")
            ffmpeg_command = [
                "ffmpeg",
                "-i",
                filepath,
                "-ss",
                str(start_time),
                "-t",
                str(segment_duration),
                "-c:v",
                "libx264",
                "-c:a",
                "aac",
                output_file,
            ]
            subprocess.run(ffmpeg_command)
        os.remove(filepath)  # remove the original video
    return


@app.route("/", methods=["POST"])
def videoSplit():
    code_start_time = time.perf_counter()

    data = request.data.decode("utf-8")
    data = json.loads(data)

    bucket_name = data["bucket"].rstrip("/")
    download_path = data["source"].rstrip("/") + "/"
    upload_path = data["destination"].rstrip("/") + "/"
    object_list = data['object_list']
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
    downloadVideos(minio_client, bucket_name, download_path, local_path, object_list=object_list)
    download_end_time = time.perf_counter()
    download_duration = download_end_time - download_start_time

    split_start_time = time.perf_counter()
    splitVideo(local_path)
    split_end_time = time.perf_counter()
    split_duration = split_end_time - split_start_time

    upload_start_time = time.perf_counter()
    uploadVideos(minio_client, bucket_name, local_path, upload_path)
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
                "split_duration": split_duration,
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
