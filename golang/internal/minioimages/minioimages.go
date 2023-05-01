package minioimages

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/minio/minio-go/v7"
)

func UploadImages(minioClient *minio.Client, bucketName string, localPath string, remotePath string) {
	filepath.Walk(localPath, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			_, mc_err := minioClient.FPutObject(context.Background(), bucketName, remotePath+info.Name(), path, minio.PutObjectOptions{ContentType: "image/jpeg"})
			if mc_err != nil {
				log.Println(mc_err)
			}
		}
		return err
	})
	log.Printf("Uploaded from %s to %s/%s\n", localPath, bucketName, remotePath)
}

func DownloadImages(minioClient *minio.Client, bucketName string, remotePath string, localPath string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: remotePath, Recursive: false})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
		}
		_, filename := filepath.Split(object.Key)
		minioClient.FGetObject(context.Background(), bucketName, object.Key, localPath+filename, minio.GetObjectOptions{})
	}
	log.Printf("Downloaded from %s/%s to %s\n", bucketName, remotePath, localPath)
}

func CreateBucket(minioClient *minio.Client, bucketName string) {
	err := minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(context.Background(), bucketName)
		if errBucketExists == nil && exists {
			log.Printf("Bucket %s exists\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Bucket %s created\n", bucketName)
	}
}

func RemoveImages(minioClient *minio.Client, bucketName string, remotePath string) {
	// Remove one by one
	objectsCh := make(chan minio.ObjectInfo)
	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for obj := range minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{Prefix: remotePath, Recursive: false}) {
			if obj.Err != nil {
				log.Fatalln(obj.Err)
			}
			objectsCh <- obj
		}
	}()
	for err := range minioClient.RemoveObjects(context.Background(), bucketName, objectsCh, minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}) {
		fmt.Println("Error detected during deletion: ", err)
	}

	log.Printf("Removed %s/%s\n", bucketName, remotePath)
}

func ForceRemove(minioClient *minio.Client, bucketName string, remotePath string) {
	// Remove whole object
	err := minioClient.RemoveObject(context.Background(), bucketName, remotePath, minio.RemoveObjectOptions{GovernanceBypass: true, ForceDelete: true})
	if err != nil {
		log.Println(err)
	}
	log.Printf("Removed %s/%s\n", bucketName, remotePath)
}
