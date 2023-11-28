import os
import shutil
import json
import time
import pickle
import re

import uuid
from wrapper import MinioWrapper as Minio
from flask import Flask, request, make_response

app = Flask(__name__)

# Minio
endpoint = os.getenv("MINIO_ENDPOINT")
access_key = os.getenv("MINIO_ACCESS_KEY")
secret_key = os.getenv("MINIO_SECRET_KEY")


def downloadTextFiles(minio_client: Minio, bucket_name, remote_path, local_path, object_list=[], remote_download=False):
    cnt = 0
    for obj in object_list:
        minio_client.fget_object(bucket_name, remote_path + obj, local_path + os.path.basename(obj), remote_download=remote_download)
        cnt += 1
    return cnt


def uploadWordCountDicts(minio_client: Minio, bucket_name, local_path, remote_path, remote_upload=False):
    for filename in os.listdir(local_path):
        filepath = os.path.join(local_path, filename)
        minio_client.fput_object(bucket_name, remote_path + filename, filepath, remote_upload=remote_upload)
    return


def WordCount(local_path):
    dic = {}
    for filename in os.listdir(local_path):
        filepath = os.path.join(local_path, filename)
        with open(filepath, 'r') as f:
            data = f.read()
            for w in re.split(r'\W+', data):
                if w not in dic:
                    dic[w] = 1
                else:
                    dic[w] += 1
    return dic


@app.route("/", methods=["POST"])
def wordCount():
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
    downloadTextFiles(minio_client, bucket_name, download_path, local_path, object_list=object_list, remote_download=True)
    download_end_time = time.perf_counter()
    download_duration = download_end_time - download_start_time

    count_start_time = time.perf_counter()
    dic = WordCount(local_path)
    count_end_time = time.perf_counter()
    count_duration = count_end_time - count_start_time

    pickle_path = "output/"+object_list[0].split(".")[0] + ".pkl"
    if os.path.exists("output"):
        shutil.rmtree("output")
    os.makedirs("output")
    pickle.dump(dic, open(pickle_path, "wb"))

    upload_start_time = time.perf_counter()
    uploadWordCountDicts(minio_client, bucket_name, "output", upload_path)
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
                "count_duration": count_duration,
                "upload_duration": upload_duration,
            }
        )
    )
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "word-count" #???
    return response


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=os.environ.get("PORT", 8080))
