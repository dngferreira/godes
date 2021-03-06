// Copyright 2013 Alex Goussiatiner. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package godes

import (
	"container/list"
	"fmt"
)

// Queue represents a FIFO or LIFO queue
type Queue struct {
	id      string
	fifo    bool
	sumTime float64
	count   int64
	qList   *list.List
	qTime   *list.List
}

// FIFOQueue represents a FIFO queue
type FIFOQueue struct {
	Queue
}

// LIFOQueue represents a LIFO queue
type LIFOQueue struct {
	Queue
}

//GetAverageTime is average elapsed time for an object in the queue
func (q *Queue) GetAverageTime() float64 {
	return q.sumTime / float64(q.count)
}

//Len returns number of objects in the queue
func (q *Queue) Len() int {
	return q.qList.Len()
}

//Place adds an object to the queue
func (q *Queue) Place(entity interface{}) {
	q.qList.PushFront(entity)
	q.qTime.PushFront(stime)
}

// Get returns an object and removes it from the queue
func (q *Queue) Get() interface{} {

	var entity interface{}
	var timeIn float64
	if q.fifo {
		entity = q.qList.Back().Value
		timeIn = q.qTime.Back().Value.(float64)
		q.qList.Remove(q.qList.Back())
		q.qTime.Remove(q.qTime.Back())
	} else {
		entity = q.qList.Front().Value
		timeIn = q.qTime.Front().Value.(float64)
		q.qList.Remove(q.qList.Front())
		q.qTime.Remove(q.qTime.Front())
	}

	q.sumTime = q.sumTime + stime - timeIn
	q.count++

	return entity
}

// NewFIFOQueue itializes the FIFO queue
func NewFIFOQueue(mid string) *FIFOQueue {
	return &FIFOQueue{Queue{fifo: true, id: mid, qList: list.New(), qTime: list.New()}}
}

// NewLIFOQueue itializes the LIFO queue
func NewLIFOQueue(mid string) *LIFOQueue {
	return &LIFOQueue{Queue{fifo: false, id: mid, qList: list.New(), qTime: list.New()}}
}

//Clear reinitiates the queue
func (q *Queue) Clear() {
	q.sumTime = 0
	q.count = 0
	q.qList.Init()
	q.qTime.Init()

}

func (q *Queue) String() string {
	return fmt.Sprintf(" Average Time=%6.3f ", q.GetAverageTime())
}
