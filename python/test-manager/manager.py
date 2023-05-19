import os
import shutil
import json

import argparse
import uuid
from flask import Flask, request, make_response

app = Flask(__name__)

parser = argparse.ArgumentParser(
    prog='Test manager',
    description='Simulates the manager',
)
parser.add_argument('-p', '--port', type=int, default=8080)
args = parser.parse_args()


@app.route('/', methods=['POST'])
def manager():
    data = request.data.decode("utf-8")
    data = json.loads(data)
    bucket_name = data['Bucket'].rstrip("/")
    object_name = data['Object'].rstrip("/")

    result = True
    if bucket_name == "images-processing" and os.path.dirname(object_name) == "images-scaled":
        result = False

    response = make_response({"Result": result})
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "test-manager"
    return response


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=args.port)
