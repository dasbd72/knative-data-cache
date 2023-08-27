package filelru

import (
	"os"
	"path/filepath"
)

type LRU struct {
	mapNode map[string]*node
	list    *list
}

// NewLRU creates a new LRU
func NewLRU() *LRU {
	return &LRU{
		mapNode: make(map[string]*node),
		list:    newList(),
	}
}

// Init initializes the LRU with files in the path recursively
func (q *LRU) Init(path string) error {
	err := filepath.Walk(path, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			q.Push(walkPath)
		}
		return nil
	})
	return err
}

// Push pushes a file to the tail of the LRU
func (q *LRU) Push(file string) error {
	if n, ok := q.mapNode[file]; ok {
		// move to tail
		err := q.list.remove(n)
		if err != nil {
			return err
		}
		err = q.list.push_back(n)
		if err != nil {
			return err
		}
	} else {
		// add to tail
		n := newNode(file)
		q.mapNode[file] = n
		err := q.list.push_back(n)
		if err != nil {
			return err
		}
	}

	return nil
}

// Pop pops a file from the front of the LRU
func (q *LRU) Pop() (string, error) {
	n, err := q.list.pop_front()
	if n == nil || err != nil {
		return "", err
	}
	delete(q.mapNode, n.file)

	return n.file, nil
}

// Empty checks if the LRU is empty
func (q *LRU) Empty() bool {
	return q.list.empty()
}

// Remove removes a file from the LRU
func (q *LRU) Remove(file string) error {
	if n, ok := q.mapNode[file]; ok {
		// remove from linked list
		err := q.list.remove(n)
		if err != nil {
			return err
		}
		// remove from map
		delete(q.mapNode, file)
	}
	return nil
}

// Size returns the size of the LRU
func (q *LRU) Size() int {
	return q.list.size
}

// FileList returns the list of files in the LRU
func (q *LRU) FileList() []string {
	var files []string
	for n := q.list.head.next; n != q.list.tail; n = n.next {
		files = append(files, n.file)
	}
	return files
}
