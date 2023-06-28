package filesQueue

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

type File struct {
	Timestamp int64
	Bucket    string
	Object    string
}

type FilesQueue struct {
	files []File
}

func (q *FilesQueue) Enqueue(timestamp int64, bucket, object string) {
	file := File{Timestamp: timestamp, Bucket: bucket, Object: object}
	q.files = append(q.files, file)
}

func (q *FilesQueue) Dequeue() File {
	file := q.files[0]
	q.files = q.files[1:]
	return file
}

// add below to main.go, remember to initialize "filesQueue" in func init()
var (
	filesQueue FilesQueue
)

func modified_handle_upload(w http.ResponseWriter, r *http.Request) {
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
	for len(filesQueue.files) > 0 {
		file := filesQueue.Dequeue()
		currentTime := time.Now().Unix()
		timeDiff := currentTime - file.Timestamp
		if timeDiff > 20 {
			err := os.Remove(utils.GetLocalPath(storagePath, req.Endpoint, req.Bucket, req.Object)) // remove the file
		} else {
			filesQueue.Enqueue(file.Timestamp, file.Bucket, file.Object) // put it back
			break
		}
	}
	// control the size of queue in 3000
	if len(filesQueue.files) < 3000 {
		res.Result = true
		filesQueue.Enqueue(time.Now().Unix(), req.Bucket, req.Object)
	} else {
		res.Result = false
	}
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
