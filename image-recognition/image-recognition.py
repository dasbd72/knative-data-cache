import os
import shutil
import json
import time
from datetime import date, datetime

import argparse
import uuid
from wrapper import MinioWrapper as Minio
from PIL import Image
import torch
from torchvision.models import resnet50, ResNet50_Weights
from torchvision import transforms
from flask import Flask, request, make_response

# http server
app = Flask(__name__)

# Arguments
parser = argparse.ArgumentParser(
    prog='Image Recognition',
    description='Runs resnet on the images',
)
parser.add_argument('-p', '--port', type=int, default=8080)
args = parser.parse_args()

# Minio
endpoint = "10.121.240.169:9000"
access_key = "LbtKL76UbWedONnd"
secret_key = "Bt0Omfh0S3ud5VEQAVR85CwinSULl3Sj"

# Model
SCRIPT_DIR = os.path.abspath(os.path.dirname(__file__))
class_idx = json.load(open(os.path.join(SCRIPT_DIR, "imagenet_class_index.json"), 'r'))
idx2label = [class_idx[str(k)][1] for k in range(len(class_idx))]
model = resnet50(weights=ResNet50_Weights.DEFAULT)


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


def inference(local_path):
    time.sleep(1)
    return []

    #  model
    model.eval()
    preprocess = transforms.Compose([
        transforms.ToTensor(),
    ])
    ret_list = []
    for filename in os.listdir(local_path):
        filepath = os.path.join(local_path, filename)
        img = Image.open(filepath)
        input_tensor = preprocess(img)
        input_batch = input_tensor.unsqueeze(0)
        output = model(input_batch)
        _, index = torch.max(output, 1)
        # The output has unnormalized scores. To get probabilities, you can run a softmax on it.
        prob = torch.nn.functional.softmax(output[0], dim=0)
        _, indices = torch.sort(output, descending=True)
        ret = idx2label[index]
        ret_list.append({filename: ret})
    return ret_list


@app.route('/', methods=['POST'])
def imageRecognition():
    code_start_time = time.perf_counter()

    data = request.data.decode("utf-8")
    data = json.loads(data)

    bucket_name = data['bucket'].rstrip("/")
    download_path = data['source'].rstrip("/") + "/"
    if 'force_remote' in data:
        force_remote = data['force_remote']
    else:
        force_remote = False
    if 'short_result' in data:
        short_result = data['short_result']
    else:
        short_result = False
    local_path = './storage/'

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
    downloadImages(minio_client, bucket_name, download_path, local_path)
    download_end_time = time.perf_counter()
    download_duration = download_end_time - download_start_time

    inference_start_time = time.perf_counter()
    pred_lst = inference(local_path)
    inference_end_time = time.perf_counter()
    inference_duration = inference_end_time - inference_start_time

    minio_client.close()

    code_end_time = time.perf_counter()
    code_duration = code_end_time - code_start_time

    # send response
    if short_result:
        pred_lst = pred_lst[:5]
    response = make_response(json.dumps({
        "predictions": pred_lst,
        "force_remote": force_remote,
        "short_result": short_result,
        "code_duration": code_duration,
        "download_duration": download_duration,
        "inference_duration": inference_duration,
        "download_post_duration": minio_client.get_download_perf()
    }))
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "image-recognition"
    return response


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=args.port)
