package main

import (
	"log"
	"os"
	"syscall"

	"github.com/dasbd72/images-processing-benchmarks/manager-go/pkg/filelru"
	"github.com/dasbd72/rfsnotify"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	etcdHost    string
	etcdClient  *clientv3.Client
)

func init() {
	// read storage path from environment variable
	storagePath = os.Getenv("STORAGE_PATH")

	// read etcd host from environment variable
	etcdHost = os.Getenv("ETCD_HOST")

	// initialize etcd client
	var err error
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints: []string{etcdHost},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	handleFileEvents()
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

	// release storage
	err = releaseStorage(0.5, lru)
	if err != nil {
		log.Fatal(err)
	}

	// handle file events
	for {
		select {
		case event := <-watcher.Events:
			err = handleFileEvent(event, lru)
			if err != nil {
				log.Println("[ERROR] handleFileEvent(event, lru): ", err)
			}

		case err := <-watcher.Errors:
			log.Println("[ERROR] event error:", err)
		}
	}
}

func handleFileEvent(event rfsnotify.Event, lru *filelru.LRU) error {
	// Ignore if file is *.hash
	if len(event.Name) > 5 && event.Name[len(event.Name)-5:] == ".hash" {
		return nil
	}

	if event.Has(rfsnotify.Write) || event.Has(rfsnotify.Create) || event.Has(rfsnotify.Chmod) {
		// push file to LRU if file is not directory
		if fs, err := os.Stat(event.Name); err == nil && !fs.IsDir() {
			err = lru.Push(event.Name)
			if err != nil {
				return err
			}
		}
	} else if event.Has(rfsnotify.Remove) {
		err := lru.Remove(event.Name)
		if err != nil {
			return err
		}
		_, err = etcdClient.Delete(etcdClient.Ctx(), event.Name)
		if err != nil {
			return err
		}
	} else {
		log.Println("[WARNING] event not expected: ", event)
	}

	// release storage
	err := releaseStorage(0.8, lru)
	if err != nil {
		return err
	}

	return nil
}

func releaseStorage(percentage float64, lru *filelru.LRU) error {
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

	if float64(used) > percentage*float64(all) {
		// if storage usage is greater than percentage, remove files
		log.Printf("Storage usage: %d/%d\n", used, all)

		// remove the least recently used file
		removedFiles := 0
		for !lru.Empty() && float64(used) > percentage*float64(all) {
			file, err := lru.Pop()
			if err != nil {
				return err
			}

			err = os.Remove(file)
			if err != nil {
				return err
			}

			_, err = etcdClient.Delete(etcdClient.Ctx(), file)
			if err != nil {
				return err
			}

			err = scan()
			if err != nil {
				return err
			}

			removedFiles++
		}
		log.Println("Removed ", removedFiles, " files")
	}

	return nil
}
