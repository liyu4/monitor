package queue

import (
	"fmt"
)

// define a node
type Node struct {
	Mail    []string
	Content string
	Header  string
}

// queue
type Queue struct {
	Data Node
	Next *Queue
}

func InitQueue() *Queue {
	q := &Queue{}
	q.Next = nil
	return q
}

func IsEmpty(q *Queue) bool {

	return q.Next == nil

}

// push
func EnQueue(q *Queue, data Node) {
	p := &Queue{}
	p.Data = data
	p.Next = nil

	if q.Next == nil {
		q.Next = p
	} else {
		s := q

		for s.Next != nil {

			s = s.Next
		}
		s.Next = p
	}

}

// pop
func DeQueue(q *Queue) Node {

	if IsEmpty(q) {
		return Node{}
	} else {
		s := q.Next

		if s.Next == nil {
			q.Next = nil
		} else {
			q.Next = s.Next

		}
		return s.Data
	}
}

// size
func Size(q *Queue) int {

	if IsEmpty(q) {
		return 0
	}

	s := q
	var size int
	for s.Next != nil {
		size++
		s = s.Next
	}

	return size
}

// print
func Print(q *Queue) {

	if IsEmpty(q) {
		return
	}

	s := q
	for s.Next != nil {

		fmt.Println(s.Data)
		s = s.Next
	}
}

// front
func Header(q *Queue) Node {

	if IsEmpty(q) {
		return Node{}
	}

	return q.Next.Data
}

// rear
func Tail(q *Queue) Node {

	if IsEmpty(q) {
		return Node{}
	}

	s := q
	for s.Next != nil {

		s = s.Next
	}

	return s.Next.Data
}
