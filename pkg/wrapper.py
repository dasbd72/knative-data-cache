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


# class Manager:
#     def __init__(self, endpoint) -> None:
#         self.exist = False
#         self.endpoint: str = endpoint
#         self.connection: socket.socket = None

#         if "STORAGE_PATH" in os.environ.keys():
#             self.storage_path = os.environ["STORAGE_PATH"]
#         else:
#             self.storage_path = None
#         logging.info(f"STORAGE_PATH: {self.storage_path}")

#         if os.path.exists(os.path.join(self.storage_path, "MANAGER_IP")):
#             with open(os.path.join(self.storage_path, "MANAGER_IP"), "r") as f:
#                 self.manager_ip = f.read()
#         else:
#             self.manager_ip = None

#         self.manager_port = os.environ.get("MANAGER_PORT")

#         if self.storage_path is not None and self.manager_ip is not None and self.manager_port is not None:
#             self.exist = True
#         logging.info(f"MANAGER: {self.manager_ip}:{self.manager_port}")

#         try:
#             self.connection = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
#             self.connection.connect((self.manager_ip, self.manager_port))
#         except:
#             logging.error("connection failed")
#             self.exist = False

#     def close(self) -> None:
#         if self.connection is not None:
#             self.connection.close()

#     def send_recv(self, data: str, bufsize: int = 128) -> str:
#         if self.connection is None:
#             logging.error("connection is None")
#             return "{}"
#         try:
#             # Send data
#             self.connection.send(data.encode("utf-8"))

#             # Receive data
#             result = self.connection.recv(bufsize).decode("utf-8")
#         except:
#             logging.error("send_recv failed")
#             return "{}"
#         else:
#             result = json.JSONDecoder().decode(result)
#             if "success" in result.keys() and result["success"]:
#                 return result["body"]
#             else:
#                 return "{}"


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
        logging.info("\n\nMinioWrapper init\n\n")

        self.endpoint: str = endpoint
        self.force_remote: bool = force_remote
        if os.path.exists("ETCD_HOST"):
            with open("ETCD_HOST", "r") as f:
                self.etcd_host = f.read()
        else:
            self.etcd_host = None
        logging.info(f"ETCD_HOST: {self.etcd_host}")

        try:
            self.etcd_client = etcd3.client(host=self.etcd_host, port=2379)
        except Exception as e:
            logging.error("Initialize etcd fail: {}".format(e))

        if "STORAGE_PATH" in os.environ.keys():
            self.storage_path = os.environ["STORAGE_PATH"]
        else:
            self.storage_path = None
        logging.info(f"STORAGE_PATH: {self.storage_path}")

        # Read data serve ip:port from storage
        if os.path.exists(os.path.join(self.storage_path, "DATA_SERVE_IP")):
            with open(os.path.join(self.storage_path, "DATA_SERVE_IP"), "r") as f:
                self.data_serve_ip = f.read()
        else:
            self.data_serve_ip = None
        logging.info(f"DATA_SERVE_IP: {self.data_serve_ip}")

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
        remote_upload=False,
    ):
        local_dst = self.get_local_path(bucket_name, object_name)

        def copy_to_local():
            if self.force_remote:
                logging.info("force remote")
                return False

            if not self.storage_path:
                logging.info("no storage path")
                return False

            # TODO: use a better way to check whether to copy to local
            if psutil.disk_usage(self.storage_path).free < os.path.getsize(file_path) * 2:
                logging.info("disk is full")
                return False

            try:
                logging.info("fput_object local {}".format(local_dst))

                os.makedirs(os.path.dirname(local_dst), exist_ok=True)
                shutil.copy(file_path, local_dst)

                save_hash_to_file(calculate_hash(local_dst), self.get_hash_file_path(local_dst))

                self.etcd_client.put(file_path, self.data_serve_ip)
                logging.info(
                    "read value from etcd:{}".format(
                        self.etcd_client.get(file_path)
                    )
                )

                return True
            except Exception as e:
                logging.error("fput_object local {} failed: {}".format(object_name, e))
                return False

        success = copy_to_local()

        if remote_upload or not success:
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
        remote_download=False,
    ):
        local_src = self.get_local_path(bucket_name, object_name)

        def copy_from_local():
            if self.force_remote or remote_download:
                return False
            if not self.storage_path:
                logging.info("no storage path")
                return False
            if not os.path.exists(local_src):
                logging.info("file not exists")
                return False

            try:
                logging.info("fget_object local {}".format(local_src))
                shutil.copy(local_src, file_path)
                if not verify_hash(file_path, self.get_hash_file_path(local_src)):
                    logging.info("incorrect hash value, file {} is corrupted.".format(object_name))
                    return False
                return True
            except Exception as e:
                logging.error("fget_object local {} failed: {}".format(object_name, e))
                return False

        def download_from_cluster():
            # TODO: download with socket connection
            if self.force_remote or remote_download:
                return False

            return False

        if copy_from_local():
            return

        if download_from_cluster():
            return

        # remote_download or force_remote or copy_from_local and download_from_cluster failed
        try:
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
