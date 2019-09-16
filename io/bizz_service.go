package io

type Service interface {
	Serve(Message)
}

type BizzService struct{}

func NewBizzService() *BizzService {
	return &BizzService{}
}

var wp IOWorkerPool

func init() {
	wp = NewIOWorkerPool(512)
}

func (b *BizzService) Serve(m Message) {
	wp.Send(m)
	//(*Channel)(unsafe.Pointer(m.Udata)).EncodeMessage(m)
}
