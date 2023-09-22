import os
from re import L
import shutil
from minio import Minio
import logging
import time
import socket
import json
import psutil
import hashlib
import etcd3

logging.basicConfig(
    level=logging.getLevelName(os.environ.get("LOG_LEVEL", "WARNING")),
    format="%(asctime)s %(filename)s:%(lineno)d %(levelname)s %(message)s",
)


class Manager:
    def __init__(self, endpoint) -> None:
        self.exist = False
        self.endpoint: str = endpoint
        self.connection: socket.socket = None

        if "STORAGE_PATH" in os.environ.keys():
            self.storage_path = os.environ["STORAGE_PATH"]
        else:
            self.storage_path = None
        logging.info(f"STORAGE_PATH: {self.storage_path}")

        if os.path.exists(os.path.join(self.storage_path, "MANAGER_IP")):
            with open(os.path.join(self.storage_path, "MANAGER_IP"), "r") as f:
                self.manager_ip = f.read()
        else:
            self.manager_ip = None

        self.manager_port = os.environ.get("MANAGER_PORT")

        if (
            self.storage_path is not None
            and self.manager_ip is not None
            and self.manager_port is not None
        ):
            self.exist = True
        logging.info(f"MANAGER: {self.manager_ip}:{self.manager_port}")

        try:
            self.connection = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.connection.connect((self.manager_ip, self.manager_port))
        except:
            logging.error("connection failed")
            self.exist = False

    def close(self) -> None:
        if self.connection is not None:
            self.connection.close()

    def send_recv(self, data: str, bufsize: int = 128) -> str:
        if self.connection is None:
            logging.error("connection is None")
            return "{}"
        try:
            # Send data
            self.connection.send(data.encode("utf-8"))

            # Receive data
            result = self.connection.recv(bufsize).decode("utf-8")
        except:
            logging.error("send_recv failed")
            return "{}"
        else:
            result = json.JSONDecoder().decode(result)
            if "success" in result.keys() and result["success"]:
                return result["body"]
            else:
                return "{}"


class MinioWrapper(Minio):
    """Inherited Wrapper"""

    def __init__(
        self,
        endpoint,
        access_key=None,
        secret_key=None,
        session_token=None,
        secure=True,
        region=None,
        http_client=None,
        credentials=None,
        force_remote=False,
    ):
        super().__init__(
            endpoint,
            access_key,
            secret_key,
            session_token,
            secure,
            region,
            http_client,
            credentials,
        )

        self.endpoint: str = endpoint
        self.force_remote: bool = force_remote
        try:
            self.etcd_client = etcd3.client(host="10.121.240.143", port=2379)
        except Exception as e:
            logging.error("Initialize etcd fail: {}".format(e))
        if "STORAGE_PATH" in os.environ.keys():
            self.storage_path = os.environ["STORAGE_PATH"]
        else:
            self.storage_path = None
        logging.info(f"STORAGE_PATH: {self.storage_path}")

    def fput_object(
        self,
        bucket_name,
        object_name,
        file_path,
        content_type="application/octet-stream",
        metadata=None,
        sse=None,
        progress=None,
        part_size=0,
        num_parallel_uploads=3,
        tags=None,
        retention=None,
        legal_hold=False,
    ):
        local_upload = True
        try:
            if self.force_remote:
                local_upload = False
                logging.info("force remote")

            # TODO: use a better way to check whether to copy to local
            disk_usage = psutil.disk_usage(self.storage_path)
            if disk_usage.free < os.path.getsize(file_path) * 2:
                local_upload = False
                logging.info("disk is full")

        except Exception as e:
            logging.error("{}".format(e))

        try:
            if local_upload:
                # copy to local
                try:
                    dst = self.get_local_path(bucket_name, object_name)
                    logging.info("fput_object local {}".format(dst))
                    os.makedirs(os.path.dirname(dst), exist_ok=True)
                    shutil.copy(file_path, dst)
                    save_hash_to_file(calculate_hash(dst), self.get_hash_file_path(dst))

                    self.etcd_client.put(file_path, os.environ.get("NODE_IP"))
                    logging.info(
                        "read value from etcd:{}".format(
                            self.etcd_client.get(file_path)
                        )
                    )

                except Exception as e:
                    logging.error("fput_object local {}".format(e))
            logging.info("fput_object {}".format(object_name))
            super().fput_object(
                bucket_name,
                object_name,
                file_path,
                content_type,
                metadata,
                sse,
                progress,
                part_size,
                num_parallel_uploads,
                tags,
                retention,
                legal_hold,
            )
        except Exception as e:
            logging.error("fput_object {} failed: {}".format(object_name, e))

    def fget_object(
        self,
        bucket_name,
        object_name,
        file_path,
        request_headers=None,
        ssec=None,
        version_id=None,
        extra_query_params=None,
        tmp_file_path=None,
        progress=None,
    ):
        local_download = True
        src = self.get_local_path(bucket_name, object_name)
        try:
            if self.force_remote:
                local_download = False
                logging.info("force remote")

            if not self.storage_path:
                local_download = False
                logging.info("no storage path")

            if not os.path.exists(src):
                local_download = False
                logging.info("file not exists")

        except Exception as e:
            logging.error("{}".format(e))

        try:
            if local_download:
                logging.info("fget_object local {}".format(src))
                shutil.copy(src, file_path)
                # print("local download time:",end="") # test
                # print(time.perf_counter() - local_download_time) # test
                if not verify_hash(file_path, self.get_hash_file_path(src)):
                    logging.info(
                        "incorrect hash value, file {} is corrupted.".format(
                            object_name
                        )
                    )
                    logging.info("fget_object {}".format(object_name))
                    super().fget_object(
                        bucket_name,
                        object_name,
                        file_path,
                        request_headers,
                        ssec,
                        version_id,
                        extra_query_params,
                        tmp_file_path,
                        progress,
                    )
            else:
                logging.info("fget_object {}".format(object_name))
                # remote_download_time = time.perf_counter() # test
                super().fget_object(
                    bucket_name,
                    object_name,
                    file_path,
                    request_headers,
                    ssec,
                    version_id,
                    extra_query_params,
                    tmp_file_path,
                    progress,
                )
                # print("remote download time:",end="") # test
                # print(time.perf_counter() - remote_download_time) # test

        except Exception as e:
            logging.error("fget_object {} failed: {}".format(object_name, e))

    """ Increased Methods """

    def get_local_path(self, bucket_name: str, object_name: str) -> str:
        return os.path.join(
            self.storage_path, self.endpoint.replace("/", "_"), bucket_name, object_name
        )

    def get_hash_file_path(self, file_path) -> str:
        try:
            parts = file_path.split(".")
            parts[-1] = "txt"
            hash_file_path = ".".join(parts)
        except Exception as e:
            logging.error("{}".format(e))
        return hash_file_path

    def get_upload_perf(self):
        return self.upload_perf

    def get_download_perf(self):
        return self.download_perf


def calculate_hash(file_path, hash_algorithm="sha256", buffer_size=65536):
    """Calculate the hash of a file."""
    try:
        hash_obj = hashlib.new(hash_algorithm)
        with open(file_path, "rb") as file:
            while chunk := file.read(buffer_size):
                hash_obj.update(chunk)
    except Exception as e:
        logging.error("calculate hash error: {}".format(e))
    return hash_obj.hexdigest()


def save_hash_to_file(hash_value, hash_file_path):
    """Save the hash value to a file."""
    try:
        with open(hash_file_path, "w") as hash_file:
            hash_file.write(hash_value)
    except Exception as e:
        logging.error("save hash file error: {}".format(e))


def verify_hash(file_path, hash_file_path, hash_algorithm="sha256"):
    """Verify if the hash of the file matches the provided hash value."""
    calculated_hash = calculate_hash(file_path, hash_algorithm)
    verifySuccess = False
    with open(hash_file_path, "r") as file:
        hash_value = file.read()
        verifySuccess = hash_value == calculated_hash
    return verifySuccess
