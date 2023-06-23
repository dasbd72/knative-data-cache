package minioclients

import (
	"sync"

	"github.com/minio/minio-go/v7"
)

type MinioClients struct {
	entries map[string]*entry
	mux     sync.Mutex
}

type entry struct {
	mc  *minio.Client
	mux sync.Mutex
}

// new minio clients
func NewMinioClients() *MinioClients {
	return &MinioClients{
		entries: make(map[string]*entry),
	}
}
