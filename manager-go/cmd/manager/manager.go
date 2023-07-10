package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"

	"github.com/dasbd72/images-processing-benchmarks/manager-go/pkg/minioclients"
	"github.com/dasbd72/images-processing-benchmarks/manager-go/pkg/utils"
)

type Request struct {
	Type string `json:"type"`
	Body string `json:"body"`
}

type Response struct {
	Success bool   `json:"success"`
	Body    string `json:"body"`
}

var (
	storagePath string
	hostIP      string
	mcs         *minioclients.MinioClients
)

func init() {
	// read storage path from environment variable
	storagePath = os.Getenv("STORAGE_PATH")

	// read host ip from environment variable
	hostIP = os.Getenv("HOST_IP")
	// write manager ip to storage
	f, err := os.Create(storagePath + "/MANAGER_IP")
	if err != nil {
		panic(err)
	}
	log.Println("HOST IP: " + hostIP)
	f.WriteString(hostIP)
	f.Close()

	// initialize minioclients
	mcs = minioclients.NewMinioClients()
}

func main() {
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Manager is running on " + hostIP + ":12345")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handle_connection(conn)
	}

}

func handle_connection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		var req Request
		var res Response
		// Read the request
		err := json.NewDecoder(reader).Decode(&req)
		if err != nil {
			log.Println(err)
			break
		}

		log.Printf("Request: %v\n", req)

		// Handle the request
		switch req.Type {
		case "create":
			handle_create(req, &res)
		case "download":
			handle_download(req, &res)
		case "upload":
			handle_upload(req, &res)
		case "backup":
			handle_backup(req, &res)
		}

		// Write the response
		json.NewEncoder(conn).Encode(res)
	}
}

func handle_create(request Request, response *Response) {
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
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		response.Success = false
		return
	}

	// Add the client
	err = mcs.AddClient(req.Endpoint, req.AccessKeyID, req.SecretAccessKey, req.SessionToken, req.UseSSL, req.Region)
	if err != nil {
		response.Success = false
		return
	}

	// Return success
	response.Success = true
	res.Result = true
	b, _ := json.Marshal(res)
	response.Body = string(b)
}

func handle_download(request Request, response *Response) {
	var req struct {
		Endpoint string `json:"endpoint"`
		Bucket   string `json:"bucket"`
		Object   string `json:"object"`
	}
	var res struct {
		Result bool `json:"result"` // true: allow local download, false: remote only
	}

	// Parse the request
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		response.Success = false
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
		response.Success = false
		return
	}

	if exist {
		// File exist in storage
		res.Result = true
	} else {
		// File does not exist in storage
		res.Result = false
	}

	// Return success
	response.Success = true
	b, _ := json.Marshal(res)
	response.Body = string(b)
}

func handle_upload(request Request, response *Response) {
	var req struct {
		Endpoint string `json:"endpoint"`
		Bucket   string `json:"bucket"`
		Object   string `json:"object"`
	}
	var res struct {
		Result bool `json:"result"` // true: allow local upload, false: remote only
	}

	// Parse the request
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		response.Success = false
		return
	}

	req.Bucket = strings.TrimRightFunc(req.Bucket, func(r rune) bool {
		return r == '/'
	})
	req.Object = strings.TrimRightFunc(req.Object, func(r rune) bool {
		return r == '/'
	})

	// TODO: more functions
	// Check if endpoint exist
	if mcs.Exist(req.Endpoint) {
		res.Result = true
	} else {
		response.Success = false
		return
	}

	// Return success
	response.Success = true
	b, _ := json.Marshal(res)
	response.Body = string(b)
}

func handle_backup(request Request, response *Response) {
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
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		response.Success = false
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
			response.Success = false
			return
		}
		if !exist {
			response.Success = false
			return
		}
		// upload to minio in background
		go mcs.Upload(req.Endpoint, req.Bucket, req.Object, localPath, req.ContentType)
	} else {
		response.Success = false
		return
	}

	// Return success
	response.Success = true
	res.Result = true
	b, _ := json.Marshal(res)
	response.Body = string(b)
}
