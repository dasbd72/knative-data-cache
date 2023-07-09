import os
import shutil
import requests
from minio import Minio
from minio.datatypes import Object as MinioObject
import logging
import time

logging.basicConfig(level=logging.getLevelName(os.environ.get("LOG_LEVEL", "WARNING")))


class Manager:
    def __init__(self, endpoint, access_key=None,
                 secret_key=None,
                 session_token=None,
                 secure=True,
                 region=None,) -> None:
        self.exist = False
        self.endpoint: str = endpoint

        if "STORAGE_PATH" in os.environ.keys():
            self.storage_path = os.environ["STORAGE_PATH"]
        else:
            self.storage_path = None
        logging.info(f"STORAGE_PATH: {self.storage_path}")

        if os.path.exists(os.path.join(self.storage_path, "MANAGER_URL")):
            with open(os.path.join(self.storage_path, "MANAGER_URL"), "r") as f:
                self.manager_url = f.read()
        else:
            self.manager_url = None
        logging.info(f"MANAGER_URL: {self.manager_url}")

        if self.manager_url:
            try:
                result = requests.post(self.manager_url + "/create", json={
                    'endpoint': endpoint,
                    'accessKey': access_key,
                    'secretKey': secret_key,
                    'sessionToken': session_token,
                    'secure': secure,
                    'region': region,
                })
            except:
                logging.error("unsuccessfully post create")
                self.exist = False
            else:
                result = result.json()
                if 'result' in result.keys():
                    self.exist = result['result']
                else:
                    self.exist = False
            logging.info(f"post create result {self.exist}")

    def local_download(self, bucket_name, object_name) -> bool:
        if not self.storage_path:
            return False
        try:
            #result = requests.post(self.manager_url + "/download", json={
                #'endpoint': self.endpoint,
                #'bucket': bucket_name,
                #'object': object_name
            #})
            dst = os.path.join(self.storage_path, bucket_name, object_name)
            result = os.path.exists(dst)
            return result
        except:
            logging.error("unsuccessfully check download path")
            return False
        # else:
        #     logging.info("successfully post download")
        #     result = result.json()
        #     if 'result' in result.keys():
        #         return result['result']
        #     else:
        #         return False

    def local_upload(self, bucket_name, object_name, file_path,
                     content_type="application/octet-stream") -> bool:
        if not self.manager_url or not self.storage_path:
            return False
        try:
            result = requests.post(self.manager_url + "/upload", json={
                'endpoint': self.endpoint,
                'bucket': bucket_name,
                'object': object_name
            })
        except:
            logging.error("unsuccessfully post upload")
            return False
        else:
            result = result.json()
            if 'result' in result.keys():
                return result['result']
            else:
                return False

    def backup(self, bucket_name, object_name, content_type="application/octet-stream") -> bool:
        if not self.manager_url or not self.storage_path:
            return False
        try:
            logging.info("trying to post backup")
            result = requests.post(self.manager_url + "/backup", json={
                'endpoint': self.endpoint,
                'bucket': bucket_name,
                'object': object_name,
                'contentType': content_type,
            })
        except:
            logging.error("unsuccessfully post backup")
            return False
        else:
            result = result.json()
            logging.info("successfully post backup")
            return True

    def get_local_path(self, bucket_name, object_name) -> str:
        return os.path.join(self.storage_path, self.endpoint.replace('/', '_'), bucket_name, object_name)


class MinioWrapper(Minio):
    """ Inherited Wrapper """

    def __init__(self, endpoint, access_key=None,
                 secret_key=None,
                 session_token=None,
                 secure=True,
                 region=None,
                 http_client=None,
                 credentials=None,
                 force_remote=False):
        super().__init__(endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials)

        self.force_remote = force_remote
        self.manager = Manager(endpoint, access_key, secret_key, session_token, secure, region)
        if self.force_remote:
            self.manager.exist = False

        self.upload_perf = 0
        self.download_perf = 0
        self.backup_perf = 0

    def fput_object(self, bucket_name, object_name, file_path,
                    content_type="application/octet-stream",
                    metadata=None, sse=None, progress=None,
                    part_size=0, num_parallel_uploads=3,
                    tags=None, retention=None, legal_hold=False):
        start = time.perf_counter()
        local_upload = self.manager.exist and self.manager.local_upload(bucket_name, object_name, file_path, content_type)
        self.upload_perf += time.perf_counter() - start
        if local_upload:
            # copy to local
            dst = self.manager.get_local_path(bucket_name, object_name)
            if not os.path.exists(os.path.dirname(dst)):
                os.makedirs(os.path.dirname(dst))
            logging.info("copy to local")
            shutil.copy(file_path, dst)
            # tell manager to backup
            start = time.perf_counter()
            success = self.manager.backup(bucket_name, object_name, content_type)
            self.backup_perf += time.perf_counter() - start
            if not success:
                logging.error("post backup failed, fallback to upload")
                super().fput_object(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold)
        else:
            # upload
            logging.info("upload to remote")
            super().fput_object(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold)

    def fget_object(self, bucket_name, object_name, file_path,
                    request_headers=None, ssec=None, version_id=None,
                    extra_query_params=None, tmp_file_path=None, progress=None):
        start = time.perf_counter()
        local_download = self.manager.exist and self.manager.local_download(bucket_name, object_name)
        self.download_perf += time.perf_counter() - start
        if local_download:
            logging.info("copy from local")
            src = self.manager.get_local_path(bucket_name, object_name)
            shutil.copy(src, file_path)
        else:
            logging.info("download from remote")
            super().fget_object(bucket_name, object_name, file_path, request_headers, ssec, version_id, extra_query_params, tmp_file_path, progress)

    def list_objects(self, bucket_name, prefix=None, recursive=False,
                     start_after=None, include_user_meta=False,
                     include_version=False, use_api_v1=False,
                     use_url_encoding_type=True, fetch_owner=False
                     ):
        object_set = set()
        remote_objects = super().list_objects(bucket_name, prefix, recursive, start_after, include_user_meta, include_version, use_api_v1, use_url_encoding_type, fetch_owner)
        for obj in remote_objects:
            obj: MinioObject
            object_set.add(obj.object_name)
            yield obj

        if not self.force_remote and self.manager.exist:
            if os.path.exists(os.path.join(self.manager.storage_path, bucket_name, prefix)):
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

    def get_upload_perf(self):
        return self.upload_perf

    def get_download_perf(self):
        return self.download_perf

    def get_backup_perf(self):
        return self.backup_perf
