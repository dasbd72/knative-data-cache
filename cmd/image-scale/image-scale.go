package main

import (
	"flag"
	"fmt"
	"image-processing-benchmarks/internal/minioimages"
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
			log.Printf("Connected to %s\n", endpoint)
		}
		connect_duration := time.Since(connect_start)

		download_start := time.Now()
		minioimages.DownloadImages(minioClient, bucketName, downloadPath, localPath)
		download_duration := time.Since(download_start)

		scale_start := time.Now()
		scaleImages(localPath)
		scale_duration := time.Since(scale_start)

		upload_start := time.Now()
		minioimages.UploadImages(minioClient, bucketName, localPath, uploadPath)
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
