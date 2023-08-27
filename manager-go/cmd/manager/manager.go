package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/dasbd72/images-processing-benchmarks/manager-go/pkg/lru"
	"github.com/dasbd72/rfsnotify"
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
	fileLRU     lru.LRU
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
}

func main() {
	go handle_connections()
	handle_file_events()
}

func handle_connections() {
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
		}

		// Write the response
		json.NewEncoder(conn).Encode(res)
	}
}

func handle_file_events() {
	// create rfsnotify watcher
	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(storagePath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Has(rfsnotify.Write) || event.Has(rfsnotify.Create) || event.Has(rfsnotify.Chmod) {
				log.Println("modified file:", event.Name)
				fileLRU.Push(event.Name)
			} else if event.Has(rfsnotify.Remove) {
				log.Println("removed file:", event.Name)
				fileLRU.Remove(event.Name)
			} else {
				log.Println("[WARNING] other event: ", event)
			}
		case err := <-watcher.Errors:
			log.Println("[ERROR] error:", err)
		}
	}
}
