package lru

type LRU struct {
	mapNode map[string]*node
	list    *list
}

func NewQueue() *LRU {
	return &LRU{
		mapNode: make(map[string]*node),
		list:    newList(),
	}
}

func (q *LRU) Push(file string) {
	if n, ok := q.mapNode[file]; ok {
		// move to tail
		q.list.remove(n)
		q.list.push(n)
	} else {
		// add to tail
		n := newNode(file)
		q.mapNode[file] = n
		q.list.push(n)
	}
}

func (q *LRU) Pop() string {
	n := q.list.pop()
	if n == nil {
		return ""
	}
	delete(q.mapNode, n.file)

	return n.file
}

func (q *LRU) Empty() bool {
	return q.list.empty()
}

func (q *LRU) Remove(file string) {
	if n, ok := q.mapNode[file]; ok {
		// remove from linked list
		// remove from map
		q.list.remove(n)
		delete(q.mapNode, file)
	}
}
