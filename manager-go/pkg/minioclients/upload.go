package minioclients

import (
	"context"

	"github.com/minio/minio-go/v7"
)

// upload to target endpoint
func (mcs *MinioClients) Upload(endpoint string, bucketName string, objectName string, filePath string, contentType string) {
	// if the client not exists
	if _, ok := mcs.entries[endpoint]; !ok {
		panic("entry not exists")
	}

	// upload
	_, err := mcs.entries[endpoint].mc.FPutObject(context.Background(), bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		panic(err)
	}
}
