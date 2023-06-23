import os
import shutil
import requests
from minio import Minio
from minio.datatypes import Object as MinioObject
import logging

logging.basicConfig(level=logging.INFO)


class Manager:
    def __init__(self) -> None:
        if "STORAGE_PATH" in os.environ.keys():
            self.storage_path = os.environ["STORAGE_PATH"]
            logging.info(f"STORAGE_PATH: {self.storage_path}")
        else:
            self.storage_path = None

        if os.path.exists(os.path.join(self.storage_path, "MANAGER_URL")):
            with open(os.path.join(self.storage_path, "MANAGER_URL"), "r") as f:
                self.manager_url = f.read()
                logging.info(f"MANAGER_URL: {self.manager_url}")
        else:
            self.manager_url = None

    def backup(self, bucket_name, object_name, file_path,
               content_type="application/octet-stream",
               metadata=None, sse=None, progress=None,
               part_size=0, num_parallel_uploads=3,
               tags=None, retention=None, legal_hold=False, minio_init=[]) -> bool:
        if not self.manager_url or not self.storage_path:
            return False
        try:
            logging.info("trying to post backup")
            result = requests.post(self.manager_url + "/backup", json={
                'Bucket': bucket_name,
                'Object': object_name,
                'FilePath': file_path,
                'ContentType': content_type,
                'Metadata': metadata,
                'SSE': sse,
                'Progress': progress,
                'PartSize': part_size,
                'NumParallelUploads': num_parallel_uploads,
                'Tags': tags,
                'Retention': retention,
                'LegalHold': legal_hold,
                'EndPoint': minio_init[0],
                'AccessKey': minio_init[1],
                'SecretKey': minio_init[2],
                'SessionToken': minio_init[3],
                'Secure': minio_init[4],
                'Region': minio_init[5],
                'HttpClient': minio_init[6],
                'Credential': minio_init[7]
            })
        except:
            logging.error("unsuccessfully post backup")
            return False
        else:
            result = result.json()
            logging.info("successfully post backup")
            return True

    def local_download(self, bucket_name, object_name) -> bool:
        if not self.manager_url or not self.storage_path:
            return False
        try:
            result = requests.post(self.manager_url + "/download", json={'Bucket': bucket_name, 'Object': object_name})
        except:
            logging.error("unsuccessfully post download")
            return False
        else:
            logging.info("successfully post download")
            result = result.json()
            if 'Result' in result.keys():
                return result['Result']
            else:
                return False

    def local_upload(self, bucket_name, object_name, file_path,
                     content_type="application/octet-stream",
                     metadata=None, sse=None, progress=None,
                     part_size=0, num_parallel_uploads=3,
                     tags=None, retention=None, legal_hold=False, minio_init=[]) -> bool:
        if not self.manager_url or not self.storage_path:
            return False
        try:
            result = requests.post(self.manager_url + "/upload", json={
                'Bucket': bucket_name,
                'Object': object_name,
                'FilePath': file_path,
                'ContentType': content_type,
                'Metadata': metadata,
                'SSE': sse,
                'Progress': progress,
                'PartSize': part_size,
                'NumParallelUploads': num_parallel_uploads,
                'Tags': tags,
                'Retention': retention,
                'LegalHold': legal_hold,
                'EndPoint': minio_init[0],
                'AccessKey': minio_init[1],
                'SecretKey': minio_init[2],
                'SessionToken': minio_init[3],
                'Secure': minio_init[4],
                'Region': minio_init[5],
                'HttpClient': minio_init[6],
                'Credential': minio_init[7]
            })
        except:
            logging.error("unsuccessfully post upload")
            return False
        else:
            logging.info("successfully post upload")
            result = result.json()
            if 'Result' in result.keys():
                return result['Result']
            else:
                return False

    def get_local_path(self, bucket_name, object_name) -> str:
        return os.path.join(self.storage_path, bucket_name, object_name)


class MinioWrapper(Minio):
    """ Inheritted Wrapper """
    minio_init = []

    def __init__(self, endpoint, access_key=None,
                 secret_key=None,
                 session_token=None,
                 secure=True,
                 region=None,
                 http_client=None,
                 credentials=None):
        super().__init__(endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials)

        self.manager = Manager()
        self.minio_init = [endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials]

    def fput_object(self, bucket_name, object_name, file_path,
                    content_type="application/octet-stream",
                    metadata=None, sse=None, progress=None,
                    part_size=0, num_parallel_uploads=3,
                    tags=None, retention=None, legal_hold=False):
        if self.manager.local_upload(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold, self.minio_init):
            # # delete remote
            # try:
            #     stat = super().stat_object(bucket_name, object_name)
            #     super().remove_object(bucket_name, object_name)
            # except:
            #     pass
            # copy to local
            dst = self.manager.get_local_path(bucket_name, object_name)
            if not os.path.exists(os.path.dirname(dst)):
                os.makedirs(os.path.dirname(dst))
            logging.info("copy to local")
            shutil.copy(file_path, dst)
            self.manager.backup(bucket_name, object_name, dst, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold, self.minio_init)
        else:
            # # delete local
            # dst = self.manager.get_local_path(bucket_name, object_name)
            # if os.path.exists(dst):
            #     shutil.rmtree(dst)
            # upload
            logging.info("upload to remote")
            super().fput_object(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold)

    def fget_object(self, bucket_name, object_name, file_path,
                    request_headers=None, ssec=None, version_id=None,
                    extra_query_params=None, tmp_file_path=None, progress=None):
        if self.manager.local_download(bucket_name, object_name):
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

        if self.manager.storage_path:
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
