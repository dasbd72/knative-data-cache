package minioclients

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// add minio client to the map
func (mcs *MinioClients) AddClient(endpoint, accessKeyID, secretAccessKey, sessionToken string, useSSL bool, region string) error {
	// if the entry not exists
	mcs.mux.Lock()
	if _, ok := mcs.entries[endpoint]; !ok {
		// create the entry
		mcs.entries[endpoint] = &entry{}
	}
	mcs.mux.Unlock()

	// lock the entry
	mcs.entries[endpoint].mux.Lock()
	// new minio client
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, sessionToken),
		Secure: useSSL,
		Region: region,
	})
	mcs.entries[endpoint].mc = mc
	mcs.entries[endpoint].mux.Unlock()

	return err
}
