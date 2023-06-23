package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func GetLocalPath(storage_path string, endpoint string, bucket string, object string) string {
	return filepath.Join(storage_path, strings.Replace(endpoint, "/", "_", -1), bucket, object)
}

func FileExist(local_path string) (bool, error) {
	var err error
	_, err = os.Stat(local_path)
	if err == nil {
		// File exist in storage
		return true, nil
	} else if os.IsNotExist(err) {
		// File not exist in storage
		return false, nil
	} else {
		// Error
		return false, err
	}
}
