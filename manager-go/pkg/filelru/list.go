package filelru

import "errors"

// node is a node in a doubly linked list
type node struct {
	file string
	next *node
	prev *node
}

// list is a doubly linked list
type list struct {
	head *node
	tail *node
	size int
}

// create a new node with file name
func newNode(file string) *node {
	return &node{
		file: file,
		next: nil,
		prev: nil,
	}
}

// create a new doubly linked list with dummy head and tail
func newList() *list {
	head := newNode("head")
	tail := newNode("tail")
	head.next = tail
	tail.prev = head

	return &list{
		head: head,
		tail: tail,
		size: 0,
	}
}

// push a node to the tail of the list
func (l *list) push_back(n *node) error {
	if n == l.head || n == l.tail {
		return errors.New("pushing head or tail")
	}
	if n == nil {
		return errors.New("pushing nil")
	}

	n.next = l.tail
	n.prev = l.tail.prev
	l.tail.prev.next = n
	l.tail.prev = n
	l.size++

	return nil
}

// pop a node from the front of the list
func (l *list) pop_front() (*node, error) {
	if l.size == 0 {
		return nil, errors.New("popping from an empty list")
	}

	n := l.head.next
	l.head.next = n.next
	l.head.next.prev = l.head
	l.size--

	return n, nil
}

// check if the list is empty
func (l *list) empty() bool {
	return l.size == 0
}

// remove a node from the list
func (l *list) remove(n *node) error {
	if n == l.head || n == l.tail {
		return errors.New("removing head or tail")
	}

	n.prev.next = n.next
	n.next.prev = n.prev
	l.size--

	return nil
}
