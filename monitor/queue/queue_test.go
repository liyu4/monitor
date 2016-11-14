package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {
	stack := InitQueue()

	data := DeQueue(stack)

	if data.Mail == nil {
		t.Log("hello")
	}
}
