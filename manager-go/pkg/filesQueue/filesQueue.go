package filesQueue

type File struct {
	Timestamp int64
	Endpoint  string
	Bucket    string
	Object    string
}

type FilesQueue struct {
	files []File
}

func (q *FilesQueue) Enqueue(timestamp int64, endpoint, bucket, object string) {
	file := File{Timestamp: timestamp, Endpoint: endpoint, Bucket: bucket, Object: object}
	q.files = append(q.files, file)
}

func (q *FilesQueue) Dequeue() {
	q.files = q.files[1:]
}

func (q *FilesQueue) Front() File {
	return q.files[0]
}

func (q *FilesQueue) Size() int {
	return len(q.files)
}
