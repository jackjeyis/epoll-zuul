package io

import "unsafe"

type MsgType byte

type Message struct {
	Udata   *byte
	Body    []byte
	msgType MsgType
}

func (m *Message) Hash() int {
	return int((*Channel)(unsafe.Pointer(m.Udata)).Ev.Fd)
}
