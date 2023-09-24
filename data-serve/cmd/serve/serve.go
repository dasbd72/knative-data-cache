package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
)

var (
	storagePath   string
	hostIP        string
	dataServePort string
)

func init() {
	// read storage path from environment variable
	storagePath = os.Getenv("STORAGE_PATH")

	// read host ip from environment variable
	hostIP = os.Getenv("HOST_IP")
	dataServePort = os.Getenv("DATA_SERVE_PORT")
	log.Printf("IP:PORT %s:%s\n", hostIP, dataServePort)

	// Write Data Serve IP:PORT to file
	f, err := os.Create(storagePath + "/DATA_SERVE_IP")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(hostIP + ":" + dataServePort)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	handleConnections()
}

func handleConnections() {
	listener, err := net.Listen("tcp", ":"+dataServePort)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Data Server is running on " + hostIP + ":" + dataServePort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("[Error] listener.Accept(): ", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	var req Request
	reader := bufio.NewReader(conn)

	// read request from client
	err := json.NewDecoder(reader).Decode(&req)
	if err != nil {
		log.Println(err)
		return
	}

	// Handle the request
	switch req.Type {
	case "download":
		// open file
		file, err := os.Open(req.Body)
		if err != nil {
			log.Println(err)
			return
		}
		defer file.Close()

		// send file
		_, err = io.Copy(conn, file)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("File %s sent to %d", req.Body, conn.RemoteAddr())
	}

	// close connection
	conn.Close()
}
