import os
import shutil
import requests
from minio import Minio
from minio.datatypes import Object


class Manager:
    def __init__(self) -> None:
        if "MANAGER_URL" in os.environ.keys():
            self.manager_url = os.environ["MANAGER_URL"]
            print(f"MANAGER_URL: {self.manager_url}")
        else:
            self.manager_url = None
        if "STORAGE_PATH" in os.environ.keys():
            self.storage_path = os.environ["STORAGE_PATH"]
            print(f"STORAGE_PATH: {self.storage_path}")
        else:
            self.storage_path = None

    def local_download(self, bucket_name, object_name) -> bool:
        if not self.manager_url or not self.storage_path:
            return False
        try:
            result = requests.post(self.manager_url + "/download", json={'Bucket': bucket_name, 'Object': object_name})
            result = result.json()
            if 'Result' in result.keys():
                return result['Result']
        except:
            return False
        return False

    def local_upload(self, bucket_name, object_name) -> bool:
        if not self.manager_url or not self.storage_path:
            return False
        try:
            result = requests.post(self.manager_url + "/upload", json={'Bucket': bucket_name, 'Object': object_name})
            result = result.json()
            if 'Result' in result.keys():
                return result['Result']
        except:
            return False
        return False

    def get_local_path(self, bucket_name, object_name) -> str:
        return os.path.join(self.storage_path, bucket_name, object_name)


class MinioWrapper(Minio):
    """ Inheritted Wrapper """

    def __init__(self, endpoint, access_key=None,
                 secret_key=None,
                 session_token=None,
                 secure=True,
                 region=None,
                 http_client=None,
                 credentials=None):
        super().__init__(endpoint, access_key, secret_key, session_token, secure, region, http_client, credentials)

        self.manager = Manager()

    def fput_object(self, bucket_name, object_name, file_path,
                    content_type="application/octet-stream",
                    metadata=None, sse=None, progress=None,
                    part_size=0, num_parallel_uploads=3,
                    tags=None, retention=None, legal_hold=False):
        if self.manager.local_upload(bucket_name, object_name):
            # delete remote
            try:
                stat = super().stat_object(bucket_name, object_name)
                super().remove_object(bucket_name, object_name)
            except:
                pass
            # copy to local
            dst = self.manager.get_local_path(bucket_name, object_name)
            if not os.path.exists(os.path.dirname(dst)):
                os.makedirs(os.path.dirname(dst))
            shutil.copy(file_path, dst)
        else:
            # delete local
            dst = self.manager.get_local_path(bucket_name, object_name)
            if os.path.exists(dst):
                shutil.rmtree(dst)
            # upload
            super().fput_object(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold)

    def fget_object(self, bucket_name, object_name, file_path,
                    request_headers=None, ssec=None, version_id=None,
                    extra_query_params=None, tmp_file_path=None, progress=None):
        if self.manager.local_download(bucket_name, object_name):
            src = self.manager.get_local_path(bucket_name, object_name)
            shutil.copy(src, file_path)
        else:
            super().fget_object(bucket_name, object_name, file_path, request_headers, ssec, version_id, extra_query_params, tmp_file_path, progress)

    def list_objects(self, bucket_name, prefix=None, recursive=False,
                     start_after=None, include_user_meta=False,
                     include_version=False, use_api_v1=False,
                     use_url_encoding_type=True, fetch_owner=False
                     ):
        remote_objects = super().list_objects(bucket_name, prefix, recursive, start_after, include_user_meta, include_version, use_api_v1, use_url_encoding_type, fetch_owner)
        for obj in remote_objects:
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
                        yield Object(bucket_name, filename)

    """ Increased Methods """
