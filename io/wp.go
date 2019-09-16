package io

import "unsafe"

type Stage interface {
	Start()
	Wait()
	Stop()
}

type IOWorkerPool struct {
	wp   []IOWorker
	size int
	quit chan struct{}
}

func NewIOWorkerPool(size int) IOWorkerPool {
	wp := make([]IOWorker, size)
	quit := make(chan struct{})
	for i := 0; i < size; i++ {
		wp[i] = IOWorker{
			queue: make(chan Message, 512),
			quit:  quit,
		}
		go wp[i].Run()
	}
	return IOWorkerPool{
		wp:   wp,
		size: size,
		quit: quit,
	}
}

func (w *IOWorkerPool) Send(msg Message) {
	select {
	case w.wp[msg.Hash()%w.size].queue <- msg:
	default:
	}
}

type IOWorker struct {
	queue chan Message
	quit  chan struct{}
}

func (w IOWorker) Run() {
	for {
		select {
		case <-w.quit:
			return
		case m := <-w.queue:
			w.Handle(m)
		}
	}
}

func (w IOWorker) Handle(m Message) {
	(*Channel)(unsafe.Pointer(m.Udata)).EncodeMessage(m)
}
