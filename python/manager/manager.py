import shutil
from curses.ascii import isalpha
import os
from kubernetes import client, config
from flask import Flask, request, make_response
app = Flask(__name__)


def get_host_volume_usage(directory):
    # Get disk usage statistics of the specified directory on the host
    usage = shutil.disk_usage(directory)

    # Convert the usage values to human-readable format
    total_size = usage.total / (1024 ** 2)  # Total size in GB
    used_size = usage.used / (1024 ** 2)    # Used size in GB
    free_size = usage.free / (1024 ** 2)    # Free size in GB

    # Return the disk usage statistics
    return {
        'total_size': total_size,
        'used_size': used_size,
        'free_size': free_size
    }


def get_mem_usage(v1, nodeName):
    nodes = v1.list_node().items
    memory_capacity = None
    memory_allocatable = None

    for node in nodes:
        # node_name = node.metadata.name
        # if(node_name==nodeName):
        print(node.metadata.name)
        memory_capacity = node.status.capacity["memory"]
        memory_allocatable = node.status.allocatable["memory"]
    return memory_allocatable, memory_capacity


def get_pv_usage(v1, pvName):
    pv_list = v1.list_persistent_volume()
    storage_capacity = None
    storage_usage = None

    for pv in pv_list.items:
        pv_name = pv.metadata.name

        if (pv_name == pvName):
            storage_capacity = pv.spec.capacity["storage"]
            # storage_usage = pv.status.capacity["storage"]

    return storage_usage, storage_capacity


@app.route('/', methods=['POST'])
def handle_request():
    config.load_incluster_config()

    v1 = client.CoreV1Api()

    memory_allocatable, memory_capacity = get_mem_usage(v1, "")
    storage_usage, storage_capacity = get_pv_usage(v1, "shared-volume")
    directory_path = '/shared'

    # Call the function to get the host volume usage
    volume_usage = get_host_volume_usage(directory_path)

    # Print the results
    print(storage_capacity, storage_usage)
    print(memory_capacity, memory_allocatable)
    print(f"Total Size: {volume_usage['total_size']:.5f} MB")
    print(f"Used Size: {volume_usage['used_size']:.5f} MB")
    print(f"Free Size: {volume_usage['free_size']:.5f} MB")

    mem_cap = int(memory_capacity[0:-2])
    mem_aloc = int(memory_allocatable[0:-2])
    if (mem_aloc/mem_cap > 0.2 and volume_usage['free_size'] > 50):
        return 'True'
    else:
        return 'False'


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=8080)
