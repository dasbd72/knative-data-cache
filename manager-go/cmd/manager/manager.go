package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dasbd72/images-processing-benchmarks/manager-go/pkg/filesQueue" //hasn't merge to main yet
	"github.com/dasbd72/images-processing-benchmarks/manager-go/pkg/minioclients"
	"github.com/dasbd72/images-processing-benchmarks/manager-go/pkg/utils"
)

var (
	storagePath string
	hostIP      string
	mcs         *minioclients.MinioClients
	fq          filesQueue.FilesQueue
)

func init() {
	// read storage path from environment variable
	storagePath = os.Getenv("STORAGE_PATH")

	// read host ip from environment variable
	hostIP = os.Getenv("HOST_IP")
	// write manager url to storage
	f, err := os.Create(storagePath + "/MANAGER_URL")
	if err != nil {
		panic(err)
	}
	f.WriteString("http://" + hostIP + ":12345")
	f.Close()

	// initialize minioclients
	mcs = minioclients.NewMinioClients()
}

func main() {
	http.HandleFunc("/", handle_root)
	http.HandleFunc("/create", handle_create)
	http.HandleFunc("/download", handle_download)
	http.HandleFunc("/upload", handle_upload)
	http.HandleFunc("/backup", handle_backup)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handle_root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var res struct {
		Result string `json:"result"`
	}

	res.Result = "This is manager."

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handle_create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req struct {
		Endpoint        string `json:"endpoint"`
		AccessKeyID     string `json:"accessKey"`
		SecretAccessKey string `json:"secretKey"`
		SessionToken    string `json:"sessionToken"`
		UseSSL          bool   `json:"secure"`
		Region          string `json:"region"`
	}
	var res struct {
		Result bool `json:"result"`
	}

	// Parse the request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add the client
	err = mcs.AddClient(req.Endpoint, req.AccessKeyID, req.SecretAccessKey, req.SessionToken, req.UseSSL, req.Region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	res.Result = true

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handle_download(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req struct {
		Endpoint string `json:"endpoint"`
		Bucket   string `json:"bucket"`
		Object   string `json:"object"`
	}
	var res struct {
		Result bool `json:"result"` // true: allow local download, false: remote only
	}

	// Parse the request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Bucket = strings.TrimRightFunc(req.Bucket, func(r rune) bool {
		return r == '/'
	})
	req.Object = strings.TrimRightFunc(req.Object, func(r rune) bool {
		return r == '/'
	})

	// Check if file exists in storage
	exist, err := utils.FileExist(utils.GetLocalPath(storagePath, req.Endpoint, req.Bucket, req.Object))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if exist {
		// File exist in storage
		res.Result = true
	} else {
		// File does not exist in storage
		res.Result = false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handle_upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req struct {
		Endpoint string `json:"endpoint"`
		Bucket   string `json:"bucket"`
		Object   string `json:"object"`
	}
	var res struct {
		Result bool `json:"result"` // true: allow local upload, false: remote only
	}

	// Parse the request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Bucket = strings.TrimRightFunc(req.Bucket, func(r rune) bool {
		return r == '/'
	})
	req.Object = strings.TrimRightFunc(req.Object, func(r rune) bool {
		return r == '/'
	})

	// TODO: more functions
	// first we check whether there's something out-of-date (exceed 20 seconds)
	fmt.Print("start\n")
	for fq.Size() > 0 {
		file := fq.Front()
		currentTime := time.Now().Unix()
		timeDiff := currentTime - file.Timestamp
		if timeDiff > 20 {
			os.Remove(utils.GetLocalPath(storagePath, file.Endpoint, file.Bucket, file.Object)) // remove the file
			fq.Dequeue()
		} else {
			break
		}
	}
	// control the size of queue in 3000
	if fq.Size() < 3000 {
		res.Result = true
		fq.Enqueue(time.Now().Unix(), req.Endpoint, req.Bucket, req.Object)
		fmt.Printf("queue size : %d\n", fq.Size())
	} else {
		res.Result = false
	}
	fmt.Print("end\n")
	// TODO end

	// Check if endpoint exist
	if mcs.Exist(req.Endpoint) {
		//res.Result = true
	} else {
		http.Error(w, "Minio client not initialized.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handle_backup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req struct {
		Endpoint    string `json:"endpoint"`
		Bucket      string `json:"bucket"`
		Object      string `json:"object"`
		ContentType string `json:"contentType"`
	}
	var res struct {
		Result bool `json:"result"` // true: ok to backup, false: exists error
	}

	// Parse the request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Bucket = strings.TrimRightFunc(req.Bucket, func(r rune) bool {
		return r == '/'
	})
	req.Object = strings.TrimRightFunc(req.Object, func(r rune) bool {
		return r == '/'
	})

	// Check if endpoint exist
	if mcs.Exist(req.Endpoint) {
		// Check if file exists in storage
		localPath := utils.GetLocalPath(storagePath, req.Endpoint, req.Bucket, req.Object)
		exist, err := utils.FileExist(localPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exist {
			http.Error(w, "File not exist.", http.StatusInternalServerError)
			return
		}
		// upload to minio in background
		go mcs.Upload(req.Endpoint, req.Bucket, req.Object, localPath, req.ContentType)
	} else {
		http.Error(w, "Minio client not initialized.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
