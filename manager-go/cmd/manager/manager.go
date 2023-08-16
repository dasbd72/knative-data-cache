package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
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
