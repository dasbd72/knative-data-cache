import os
import shutil
import requests
from minio import Minio


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

    def is_remote(self, bucket_name, object_name) -> bool:
        if not self.manager_url or not self.storage_path:
            return True
        try:
            result = requests.post(self.manager_url, json={'Bucket': bucket_name, 'Object': object_name})
            result = result.json()
            if 'Result' in result.keys():
                return result['Result']
        except:
            return True
        return True

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
        if self.manager.is_remote(bucket_name, object_name):
            super().fput_object(bucket_name, object_name, file_path, content_type, metadata, sse, progress, part_size, num_parallel_uploads, tags, retention, legal_hold)
        else:
            dst = self.manager.get_local_path(bucket_name, object_name)
            if not os.path.exists(os.path.dirname(dst)):
                os.makedirs(os.path.dirname(dst))
            shutil.copy(file_path, dst)

    def fget_object(self, bucket_name, object_name, file_path,
                    request_headers=None, ssec=None, version_id=None,
                    extra_query_params=None, tmp_file_path=None, progress=None):
        if self.manager.is_remote(bucket_name, object_name):
            super().fget_object(bucket_name, object_name, file_path, request_headers, ssec, version_id, extra_query_params, tmp_file_path, progress)
        else:
            src = self.manager.get_local_path(bucket_name, object_name)
            shutil.copy(src, file_path)

    def list_objects(self, bucket_name, prefix=None, recursive=False,
                     start_after=None, include_user_meta=False,
                     include_version=False, use_api_v1=False,
                     use_url_encoding_type=True, fetch_owner=False
                     ):
        lst = super().list_objects(bucket_name, prefix, recursive, start_after, include_user_meta, include_version, use_api_v1, use_url_encoding_type, fetch_owner)
        try:
            if prefix:
                lst.append(os.listdir(os.path.join(self.local_storage_path, bucket_name, prefix)))
            else:
                lst.append(os.listdir(os.path.join(self.local_storage_path, bucket_name)))
        except:
            pass
        return lst

    """ Increased Methods """
