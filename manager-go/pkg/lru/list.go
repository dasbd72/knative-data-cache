package lru

import "errors"

type node struct {
	file string
	next *node
	prev *node
}

type list struct {
	head *node
	tail *node
	size int
}

func newNode(file string) *node {
	return &node{
		file: file,
		next: nil,
		prev: nil,
	}
}

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

func (l *list) push(n *node) {
	n.next = l.tail
	n.prev = l.tail.prev
	l.tail.prev.next = n
	l.tail.prev = n
	l.size++
}

func (l *list) pop() *node {
	if l.size == 0 {
		return nil
	}

	n := l.head.next
	l.head.next = n.next
	l.head.next.prev = l.head
	l.size--

	return n
}

func (l *list) empty() bool {
	return l.size == 0
}

func (l *list) remove(n *node) error {
	if n == l.head || n == l.tail {
		return errors.New("removing head or tail")
	}

	n.prev.next = n.next
	n.next.prev = n.prev
	l.size--
	return nil
}
