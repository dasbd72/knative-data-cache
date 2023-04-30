package main

import (
	"context"
	"flag"
	"fmt"
	"image/jpeg"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nfnt/resize"
)

const (
	endpoint        = "10.121.240.169:9000"              // without http://
	accessKeyID     = "LbtKL76UbWedONnd"                 //
	secretAccessKey = "Bt0Omfh0S3ud5VEQAVR85CwinSULl3Sj" // secret key from minio console
	useSSL          = false                              // no certificate
)

var (
	address string
)

func init() {
	flag.StringVar(&address, "address", "0.0.0.0:9090", "The address of host.")
}

func main() {
	flag.Parse()

	log.Printf("Server address: %s\n", address)

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(address, nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprintf(w, "Not For You\n")
	case "POST":
		bucketName := "images-processing"                       // minio bucket name
		downloadPath := "images/"                               // source image path
		uploadPath := time.Now().Format("20060102150405") + "/" // output image path
		localPath := "storage/"

		connect_start := time.Now()
		// Initialize minio client object.
		minioClient, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			log.Fatalln(err)
		} else {
			// minioClient is now set up
			log.Printf("Connected to %s\n", endpoint)
		}
		connect_duration := time.Since(connect_start)

		download_start := time.Now()
		downloadImages(minioClient, bucketName, downloadPath, localPath)
		download_duration := time.Since(download_start)

		scale_start := time.Now()
		scaleImages(localPath)
		scale_duration := time.Since(scale_start)

		upload_start := time.Now()
		uploadImages(minioClient, bucketName, localPath, uploadPath)
		upload_duration := time.Since(upload_start)

		log.Println("Connect:", connect_duration)
		log.Println("Download:", download_duration)
		log.Println("Scale:", scale_duration)
		log.Println("Upload:", upload_duration)

		fmt.Fprintf(w, "Connect: %s\n", connect_duration)
		fmt.Fprintf(w, "Download: %s\n", download_duration)
		fmt.Fprintf(w, "Scale: %s\n", scale_duration)
		fmt.Fprintf(w, "Upload: %s\n", upload_duration)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func scaleImages(localPath string) {
	cnt := 0
	filepath.Walk(localPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		// decode jpeg into image.Image
		img, err := jpeg.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		// resize to width 1000 using Lanczos resampling and preserve aspect ratio
		m := resize.Resize(1000, 0, img, resize.Lanczos3)

		out, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		// write new image to file
		err = jpeg.Encode(out, m, &jpeg.Options{Quality: 100})
		if err != nil {
			log.Fatal(err)
		}
		cnt++

		return nil
	})
	log.Printf("Scaled %d images\n", cnt)
}

func createBucket(minioClient *minio.Client, bucketName string) {
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

func downloadImages(minioClient *minio.Client, bucketName string, remotePath string, localPath string) {
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

func uploadImages(minioClient *minio.Client, bucketName string, localPath string, remotePath string) {
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

func removeImages(minioClient *minio.Client, bucketName string, remotePath string) {
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

func forceRemove(minioClient *minio.Client, bucketName string, remotePath string) {
	// Remove whole object
	err := minioClient.RemoveObject(context.Background(), bucketName, remotePath, minio.RemoveObjectOptions{GovernanceBypass: true, ForceDelete: true})
	if err != nil {
		log.Println(err)
	}
	log.Printf("Removed %s/%s\n", bucketName, remotePath)
}
