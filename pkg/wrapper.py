import os
import shutil
import requests
from minio import Minio
from minio.datatypes import Object as MinioObject
import logging
import time
import socket
import json
import psutil

logging.basicConfig(level=logging.getLevelName(os.environ.get("LOG_LEVEL", "WARNING")))


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

    def create(
        self,
        endpoint,
        access_key=None,
        secret_key=None,
        session_token=None,
        secure=True,
        region=None,
    ):
        if not self.manager_ip or not self.storage_path:
            return False
        try:
            result = self.send_recv(
                json.JSONEncoder().encode(
                    {
                        "type": "create",
                        "body": json.JSONEncoder().encode(
                            {
                                "endpoint": endpoint,
                                "accessKey": access_key,
                                "secretKey": secret_key,
                                "sessionToken": session_token,
                                "secure": secure,
                                "region": region,
                            }
                        ),
                    }
                )
            )
        except:
            logging.error("unsuccessfully send create")
            self.exist = False
        else:
            logging.info("successfully send create")
            result = json.JSONDecoder().decode(result)
            if "result" in result.keys():
                self.exist = result["result"]
            else:
                self.exist = False

    def local_download(self, bucket_name, object_name) -> bool:
        if not self.manager_ip or not self.storage_path:
            return False
        try:
            dst = os.path.join(
                self.storage_path,
                self.endpoint.replace("/", "_"),
                bucket_name,
                object_name,
            )
            result = os.path.exists(dst)
            return result

        except:
            logging.error("unsuccessfully check whether the file exists")
            return False

    def local_upload(
        self,
        bucket_name,
        object_name,
        file_path,
        content_type="application/octet-stream",
    ) -> bool:
        if not self.manager_ip or not self.storage_path:
            return False
        try:
            result = True
            disk_usage = psutil.disk_usage(self.storage_path)
            # print(f"{disk_usage.used}, {disk_usage.percent}")
            if disk_usage.percent > 90:
                result = False  # if already used 90% of memory
            return result
        except:
            logging.error("unsuccessfully check disk usage")
            return False

    def backup(
        self, bucket_name, object_name, content_type="application/octet-stream"
    ) -> bool:
        if not self.manager_ip or not self.storage_path:
            return False
        try:
            result = self.send_recv(
                json.JSONEncoder().encode(
                    {
                        "type": "backup",
                        "body": json.JSONEncoder().encode(
                            {
                                "endpoint": self.endpoint,
                                "bucket": bucket_name,
                                "object": object_name,
                                "contentType": content_type,
                            }
                        ),
                    }
                )
            )
        except:
            logging.error("unsuccessfully send backup")
            return False
        else:
            result = json.JSONDecoder().decode(result)
            logging.info("successfully send backup")
            return True

    def get_local_path(self, bucket_name, object_name) -> str:
        return os.path.join(
            self.storage_path, self.endpoint.replace("/", "_"), bucket_name, object_name
        )


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
        force_backup=False,
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

        self.force_remote = force_remote
        self.manager = Manager(endpoint)
        self.force_backup = force_backup
        if self.force_remote:
            self.manager.exist = False
        if self.manager.exist:
            self.manager.create(
                endpoint, access_key, secret_key, session_token, secure, region
            )
        logging.info(f"manager exist: {self.manager.exist}")

        self.upload_perf = 0
        self.download_perf = 0
        self.backup_perf = 0

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
        start = time.perf_counter()
        local_upload = self.manager.exist and self.manager.local_upload(
            bucket_name, object_name, file_path, content_type
        )
        self.upload_perf += time.perf_counter() - start
        try:
            if local_upload:
                # copy to local
                dst = self.manager.get_local_path(bucket_name, object_name)
                os.makedirs(os.path.dirname(dst))
                logging.info("copy to local:{}".format(dst))
                shutil.copy(file_path, dst)
                # tell manager to backup
                if self.force_backup:
                    logging.info("force backup {}".format(dst))
                    start = time.perf_counter()
                    success = self.manager.backup(bucket_name, object_name, content_type)
                    self.backup_perf += time.perf_counter() - start
                    if not success:
                        logging.error("post backup failed, fallback to upload")
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
            else:
                # upload
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
        start = time.perf_counter()
        local_download = self.manager.exist and self.manager.local_download(
            bucket_name, object_name
        )
        self.download_perf += time.perf_counter() - start
        try:
            if local_download:
                src = self.manager.get_local_path(bucket_name, object_name)
                logging.info("copying from local:{}".format(src))
                shutil.copy(src, file_path)
                os.remove(src)  # delete file after use once
                logging.info("remove local file {}".format(src))
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

    def list_objects(
        self,
        bucket_name,
        prefix=None,
        recursive=False,
        start_after=None,
        include_user_meta=False,
        include_version=False,
        use_api_v1=False,
        use_url_encoding_type=True,
        fetch_owner=False,
    ):
        object_set = set()
        remote_objects = super().list_objects(
            bucket_name,
            prefix,
            recursive,
            start_after,
            include_user_meta,
            include_version,
            use_api_v1,
            use_url_encoding_type,
            fetch_owner,
        )
        for obj in remote_objects:
            obj: MinioObject
            object_set.add(obj.object_name)
            yield obj

        if not self.force_remote and self.manager.exist:
            if os.path.exists(
                os.path.join(self.manager.storage_path, bucket_name, prefix)
            ):
                bucket_dir = os.path.join(self.manager.storage_path, bucket_name)
                if prefix.endswith("/"):
                    object_dir = os.path.normpath(prefix)
                    files = os.listdir(os.path.join(bucket_dir, object_dir))
                else:
                    object_dir = os.path.normpath(os.path.dirname(prefix))
                    files = os.listdir(os.path.join(bucket_dir, object_dir))
                for file in files:
                    filename = os.path.normpath(os.path.join(object_dir, file))
                    if os.path.isdir(os.path.join(bucket_dir, filename)):
                        filename += "/"
                    if filename.startswith(prefix):
                        if filename not in object_set:
                            yield MinioObject(bucket_name, filename)

    """ Increased Methods """

    def close(self):
        self.manager.close()

    def get_upload_perf(self):
        return self.upload_perf

    def get_download_perf(self):
        return self.download_perf

    def get_backup_perf(self):
        return self.backup_perf
