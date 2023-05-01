package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/jpeg"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dasbd72/images-processing-benchmarks/golang/internal/minioimages"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nfnt/resize"
)

const (
	localPath       = "storage/"                         // path of local storage
	endpoint        = "10.121.240.169:9000"              // without http://
	accessKeyID     = "LbtKL76UbWedONnd"                 //
	secretAccessKey = "Bt0Omfh0S3ud5VEQAVR85CwinSULl3Sj" // secret key from minio console
	useSSL          = false                              // no certificate
)

var (
	port int  // port of host
	dry  bool // dry run
)

func init() {
	flag.IntVar(&port, "port", 9090, "Port of server")
	flag.BoolVar(&dry, "dry", false, "Dryrun")
}

func main() {
	flag.Parse()

	log.Printf("Server port: %d\n", port)

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprintf(w, "Not For You\n")
	case "POST":
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		var request struct {
			Bucket string // Bucket name
			Source string // Path of image directory
			Width  int
			Height int
		}
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		bucketName := filepath.Clean(request.Bucket)
		downloadPath := filepath.Clean(request.Source) + "/"
		uploadPath := filepath.Clean(request.Source) + "-" + time.Now().Format("20060102150405") + "/"

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
		minioimages.CreateBucket(minioClient, bucketName)
		connect_duration := time.Since(connect_start)

		download_start := time.Now()
		if !dry {
			minioimages.DownloadImages(minioClient, bucketName, downloadPath, localPath)
		}
		download_duration := time.Since(download_start)

		scale_start := time.Now()
		if !dry {
			scaleImages(localPath, request.Width, request.Height)
		}
		scale_duration := time.Since(scale_start)

		upload_start := time.Now()
		if !dry {
			minioimages.UploadImages(minioClient, bucketName, localPath, uploadPath)
		}
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

func scaleImages(localPath string, width int, height int) {
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

		m := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

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
