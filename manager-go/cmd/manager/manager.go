package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/dasbd72/images-processing-benchmarks/manager-go/pkg/filelru"
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
	go handleConnections()
	handleFileEvents()
}

func handleConnections() {
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

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
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

func handleFileEvents() {
	// initialize watcher
	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(storagePath)
	if err != nil {
		log.Fatal(err)
	}

	// initialize file LRU
	lru := filelru.NewLRU()
	err = lru.Init(storagePath)
	if err != nil {
		log.Fatal(err)
	}

	// handle file events
	for {
		log.Println("Waiting for file events...")
		select {
		case event := <-watcher.Events:
			log.Println("event: ", event)
			err = handleFileEvent(event, lru)
			if err != nil {
				log.Fatal(err)
			}

		case err := <-watcher.Errors:
			log.Println("[ERROR] event error:", err)
		}
	}
}

func handleFileEvent(event rfsnotify.Event, lru *filelru.LRU) error {
	if event.Has(rfsnotify.Write) || event.Has(rfsnotify.Create) || event.Has(rfsnotify.Chmod) {
		log.Println("event modified file:", event.Name)
		err := lru.Push(event.Name)
		if err != nil {
			return err
		}
	} else if event.Has(rfsnotify.Remove) {
		log.Println("event removed file:", event.Name)
		err := lru.Remove(event.Name)
		if err != nil {
			return err
		}
	} else {
		log.Println("[WARNING] other event: ", event)
	}

	// check storage usage
	var (
		fs   syscall.Statfs_t
		all  uint64
		free uint64
		used uint64
	)

	scan := func() error {
		err := syscall.Statfs(storagePath, &fs)
		if err != nil {
			return err
		}

		all = fs.Blocks * uint64(fs.Bsize)
		free = fs.Bfree * uint64(fs.Bsize)
		used = all - free

		return nil
	}

	// first scan
	err := scan()
	if err != nil {
		return err
	}
	log.Printf("Storage usage: %d/%d\n", used, all)

	// remove the least recently used file
	for !lru.Empty() && float64(used) > 0.8*float64(all) {
		file, err := lru.Pop()
		if err != nil {
			return err
		}
		log.Println("Removing file: ", file)

		err = os.Remove(file)
		if err != nil {
			return err
		}

		err = scan()
		if err != nil {
			return err
		}
	}

	return nil
}
