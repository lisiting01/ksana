package scheduler

import (
	"container/heap"
	"ksana-service/internal/model"
	"time"
)

type JobItem struct {
	Job     *model.Job
	RunTime time.Time
	Index   int
}

type JobHeap []*JobItem

func (h JobHeap) Len() int { return len(h) }

func (h JobHeap) Less(i, j int) bool {
	return h[i].RunTime.Before(h[j].RunTime)
}

func (h JobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *JobHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*JobItem)
	item.Index = n
	*h = append(*h, item)
}

func (h *JobHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*h = old[0 : n-1]
	return item
}

func (h *JobHeap) update(item *JobItem, runTime time.Time) {
	item.RunTime = runTime
	heap.Fix(h, item.Index)
}

func (h *JobHeap) remove(item *JobItem) {
	heap.Remove(h, item.Index)
}

func NewJobHeap() *JobHeap {
	h := &JobHeap{}
	heap.Init(h)
	return h
}