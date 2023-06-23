import os
import shutil
import json
import threading
import uuid
from flask import Flask, request, make_response
from kubernetes import client, config
from minio import Minio

dbg = False
app = Flask(__name__)
databaseClient = None
storage_path = ""


def parallel_upload(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold,
                    endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials):
    """
    usage: define an upload function for manager to fork a new thread to upload
    input for data storage: bucket_name, object_name, file_path
    input for minio init : content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold, endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials
    output: None
    """

    databaseClient = Minio(endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials)
    databaseClient.fput_object(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold)
    if (dbg):
        print("successfully upload parallelly to remote by manager")


@app.route('/download', methods=['POST'])
def handle_download_request():
    """
    usage: check current shared directory to tell whether targer file is stored locally
    input: Bucket and Object
    output: True or False
    """

    data = request.data.decode("utf-8")
    data = json.loads(data)
    # storage_path = data['STORAGE_PATH'].rstrip("/")
    bucket_name = data['Bucket'].rstrip("/")
    object_name = data['Object'].rstrip("/")

    # "shared" should replace by global var
    dst = os.path.join("shared", bucket_name, object_name)
    result = False
    if os.path.exists(dst):
        result = True

    response = make_response({"Result": result})
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "test-manager"
    return response


@app.route('/upload', methods=['POST'])
def handle_upload_request():
    """
    usage: upload file to local. if local storage is full, delete not use file by FIFO
    input: Bucket and Object
    output: True or False(only when encounter error)
    TODO: stateful stack structure
    """

    data = request.data.decode("utf-8")
    data = json.loads(data)
    # config.load_incluster_config()
    # v1 = client.CoreV1Api()

    # memory_allocatable, memory_capacity = get_mem_usage(v1, "")
    # storage_usage, storage_capacity = get_pv_usage(v1, "shared-volume")
    # directory_path = '/shared'

    # # Call the function to get the host volume usage
    # volume_usage = get_host_volume_usage(directory_path)

    # # Print the results
    # print(storage_capacity, storage_usage)
    # print(memory_capacity, memory_allocatable)
    # print(f"Total Size: {volume_usage['total_size']:.5f} MB")
    # print(f"Used Size: {volume_usage['used_size']:.5f} MB")
    # print(f"Free Size: {volume_usage['free_size']:.5f} MB")

    # mem_cap = int(memory_capacity[0:-2])
    # mem_aloc = int(memory_allocatable[0:-2])

    # result = False
    # if (mem_aloc/mem_cap > 0.2 and volume_usage['free_size'] > 50):
    #     result = True
    result = True
    response = make_response({"Result": result})
    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "test-manager"
    return response


@app.route('/backup', methods=['POST'])
def handle_backup_request():
    """
    usage: backup file in remote storage parallelly
    input for data storage: bucket_name, object_name, file_path
    input for minio init : content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold, endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials
    output: True or False(only when encounter error)
    """

    data = request.data.decode("utf-8")
    data = json.loads(data)
    bucket_name = data['Bucket'].rstrip("/")
    object_name = data['Object'].rstrip("/")
    file_path = data['FilePath']
    content_type = data['ContentType']
    metadata = data['Metadata']
    sse = data['SSE']
    progress = data['Progress']
    part_size = data['PartSize']
    num_parallel_uploads = data['NumParallelUploads']
    tags = data['Tags']
    retention = data['Retention']
    legal_hold = data['LegalHold']

    endpoint = data['EndPoint']
    access_key = data['AccessKey']
    secret_key = data['SecretKey']
    session_token = data['SessionToken']
    secure = data['Secure']
    region = data['Region']
    http_client = data['HttpClient']
    credentials = data['Credential']

    if (dbg):
        print("trying parallel_upload")
    thread = threading.Thread(target=parallel_upload, args=(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold, endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials), daemon=True)
    thread.start()

    response = make_response({"Result": True})

    response.headers["Content-Type"] = "application/json"
    response.headers["Ce-Id"] = str(uuid.uuid4())
    response.headers["Ce-specversion"] = "0.3"
    response.headers["Ce-Source"] = "test-manager"
    return response


if __name__ == '__main__':
    if "STORAGE_PATH" in os.environ.keys():
        storage_path = os.environ["STORAGE_PATH"]
        print(f"STORAGE_PATH: {storage_path}")

        # write host ip to file
        host_ip = os.environ["HOST_IP"]
        print(f"HOST_IP: {host_ip}")
        with open(os.path.join(storage_path, "MANAGER_URL"), "w") as f:
            f.write(f"http://{host_ip}:8080")
    else:
        storage_path = None
    app.run(debug=True, host='0.0.0.0', port=8080)


# def get_host_volume_usage(directory):
#     # Get disk usage statistics of the specified directory on the host
#     usage = shutil.disk_usage(directory)

#     # Convert the usage values to human-readable format
#     total_size = usage.total / (1024 ** 2)  # Total size in GB
#     used_size = usage.used / (1024 ** 2)    # Used size in GB
#     free_size = usage.free / (1024 ** 2)    # Free size in GB

#     # Return the disk usage statistics
#     return {
#         'total_size': total_size,
#         'used_size': used_size,
#         'free_size': free_size
#     }


# def get_mem_usage(v1, nodeName):
#     nodes = v1.list_node().items
#     memory_capacity = None
#     memory_allocatable = None

#     for node in nodes:
#         # node_name = node.metadata.name
#         # if(node_name==nodeName):
#         print(node.metadata.name)
#         memory_capacity = node.status.capacity["memory"]
#         memory_allocatable = node.status.allocatable["memory"]
#     return memory_allocatable, memory_capacity


# def get_pv_usage(v1, pvName):
#     pv_list = v1.list_persistent_volume()
#     storage_capacity = None
#     storage_usage = None

#     for pv in pv_list.items:
#         pv_name = pv.metadata.name

#         if (pv_name == pvName):
#             storage_capacity = pv.spec.capacity["storage"]
#             # storage_usage = pv.status.capacity["storage"]

#     return storage_usage, storage_capacity

# @app.route('/init', methods=['POST'])
# def init():
#     global databaseClient
#     data = request.data.decode("utf-8")
#     data = json.loads(data)
#     endpoint = data['Endpoint']
#     access_key = data['AccessKey']
#     secret_key = data['SecretKey']
#     session_token = data['SessionToken']
#     secure = data['Secure']
#     region1 = data['Region']
#     http_client = data['HttpClient']
#     credentials = data['Credentials']
#     databaseClient = Minio(endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials)
