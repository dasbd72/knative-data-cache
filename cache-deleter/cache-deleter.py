import os
import datetime
import time


def delete_cache(folder_path):
    current_time = datetime.datetime.now()

    for root, dirs, files in os.walk(folder_path):
        for filename in files:
            if(filename == "MANAGER_IP"):
                continue
            file_path = os.path.join(root, filename)

            creation_time = datetime.datetime.fromtimestamp(
                os.path.getctime(file_path))

            time_difference = current_time - creation_time

            if time_difference.total_seconds() > 86400: # TODO: modify this time
                os.remove(file_path)
                print(f"remove file: {file_path}")


folder_path = r"/shared"

delete_cache(folder_path)
