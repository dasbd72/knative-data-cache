import os
import shutil
from minio import Minio
import logging
import time
import socket
import json
import psutil

logging.basicConfig(level=logging.getLevelName(os.environ.get("LOG_LEVEL", "WARNING")), format='%(asctime)s %(filename)s:%(lineno)d %(levelname)s %(message)s')


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
        self.manager_port = 12345

        if self.storage_path is not None and self.manager_ip is not None:
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
        local_upload = False
        try:
            if self.force_remote:
                raise Exception("force remote")

            # TODO: use a better way to check whether to copy to local
            disk_usage = psutil.disk_usage(self.storage_path)
            if disk_usage.free < os.path.getsize(file_path) * 2:
                raise Exception("disk is full")

        except Exception as e:
            pass

        else:
            local_upload = True

        try:
            if local_upload:
                # copy to local
                dst = self.get_local_path(bucket_name, object_name)
                logging.info("fput_object local {}".format(dst))
                os.makedirs(os.path.dirname(dst), exist_ok=True)
                shutil.copy(file_path, dst)
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
            logging.error("fput_object {} failed".format(object_name))
            logging.error(e)

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
        local_download = False
        try:
            if self.force_remote:
                raise Exception("force remote")

            if not self.storage_path:
                raise Exception("no storage path")

            src = self.get_local_path(bucket_name, object_name)
            if not os.path.exists(src):
                raise Exception("file not exists")

        except Exception as e:
            pass

        else:
            local_download = True

        try:
            if local_download:
                logging.info("fget_object local {}".format(src))
                src = self.get_local_path(bucket_name, object_name)
                shutil.copy(src, file_path)
            else:
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
            logging.error("fget_object {} failed".format(object_name))
            logging.error(e)

    """ Increased Methods """

    def get_local_path(self, bucket_name: str, object_name: str) -> str:
        return os.path.join(
            self.storage_path, self.endpoint.replace("/", "_"), bucket_name, object_name
        )
