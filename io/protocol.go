package io

import (
	"bytes"
	"errors"
)

type State byte
type Method byte

const (
	sw_start State = iota
	sw_method
	sw_spaces_before_uri
	sw_after_slash_in_uri
	sw_spaces_before_http
	sw_http_H
	sw_http_HT
	sw_http_HTT
	sw_http_HTTP
	sw_first_major_digit
	sw_major_digit
	sw_first_minor_digit
	sw_minor_digit
	sw_almost_done
	sw_header_start
	sw_header_name
	sw_space_before_value
	sw_header_value
	sw_space_after_value
	sw_almost_header_done
	sw_almost_req_done
)

const (
	CR = '\r'
	LF = '\n'
)

var (
	invalidMethod = errors.New("invalid Method")
	invalidReq    = errors.New("invalid Request")
	ErrHeaderLen  = errors.New("unexpected header length!")
	noData        = errors.New("not enough data!")
)

func cmp3(bs []byte, b1, b2, b3 byte) bool {
	return bs[0] == b1 && bs[1] == b2 && bs[2] == b3
}

type Protocol interface {
	Encode(Message, *IOBuffer) error
	Decode(*IOBuffer) (Message, error)
}

type HttpProtocol struct {
	uri    []byte
	header map[string][]byte
}

func NewHTTPProtocol() *HttpProtocol {
	return &HttpProtocol{}
}

func (h *HttpProtocol) Encode(msg Message, buf *IOBuffer) error {
	buf.Write([]byte("HTTP/1.0 "))
	buf.Write([]byte("200 "))
	buf.Write([]byte("OK"))
	buf.Write([]byte("\r\n"))
	buf.Write([]byte("Connection: Keep-Alive\r\n"))
	//	buf.Write([]byte("Connection: Close\r\n"))
	buf.Write([]byte("Content-Type: text/html\r\n"))
	buf.Write([]byte("Content-Length: 11\r\n"))
	buf.Write([]byte("\r\n"))
	buf.Write([]byte("Hello world"))
	return nil
}

func (h *HttpProtocol) Decode(buf *IOBuffer) (Message, error) {
	var (
		ch     byte
		name   []byte
		header = make(map[string][]byte)
	)
	state := sw_start
	for i := uint64(0); i < buf.GetReadSize(); i++ {
		ch = buf.Byte(i)
		switch state {
		case sw_start:
			if ch == CR || ch == LF {
				break
			}
			if ch < 'A' || ch > 'Z' {
				return Message{}, invalidMethod
			}
			state = sw_method
		case sw_method:
			if ch == ' ' {
				if cmp3(buf.Read(i), 'G', 'E', 'T') {
					state = sw_spaces_before_uri
					i = 0
					break
				} else {
					return Message{}, invalidMethod
				}
			}
			if ch < 'A' || ch > 'Z' {
				return Message{}, invalidMethod
			}
		case sw_spaces_before_uri:
			if ch == '/' {
				state = sw_after_slash_in_uri
				break
			}
			//log.Printf("before uri")
			return Message{}, invalidReq
		case sw_after_slash_in_uri:
			if ch == ' ' {
				h.uri = buf.Read(i)
				//log.Printf("uri %s", string(h.uri))
				state = sw_spaces_before_http
				i = 0
			}
		case sw_spaces_before_http:
			if ch == 'H' {
				state = sw_http_H
			} else {
				//log.Printf("before h")
				return Message{}, invalidReq
			}
		case sw_http_H:
			if ch == 'T' {
				state = sw_http_HT
			} else {
				//log.Printf("before ht")
				return Message{}, invalidReq
			}
		case sw_http_HT:
			if ch == 'T' {
				state = sw_http_HTT
			} else {
				//log.Printf("before htt")
				return Message{}, invalidReq
			}
		case sw_http_HTT:
			if ch == 'P' {
				state = sw_http_HTTP
			} else {
				//log.Printf("before http")
				return Message{}, invalidReq
			}
		case sw_http_HTTP:
			if ch == '/' {
				state = sw_first_major_digit
			} else {
				//log.Printf("before major")
				return Message{}, invalidReq
			}
		case sw_first_major_digit:
			if ch < '1' || ch > '9' {
				//log.Printf("before major digit")
				return Message{}, invalidReq
			}
			state = sw_major_digit
		case sw_major_digit:
			if ch != '.' {
				//log.Printf("before digit")
				return Message{}, invalidReq
			}
			state = sw_first_minor_digit
		case sw_first_minor_digit:
			//log.Printf("before minor %d", ch)
			if ch < '0' || ch > '9' {
				return Message{}, invalidReq
			}
			state = sw_minor_digit
		case sw_minor_digit:
			if ch != CR {
				//log.Printf("before minor digit")
				return Message{}, invalidReq
			}
			buf.Read(i)
			state = sw_almost_done
			i = 0
		case sw_almost_done:
			if ch != LF {

				//log.Printf("before almost done")
				return Message{}, invalidReq
			}
			buf.Consume(2)
			i = 0
			state = sw_header_start
		case sw_header_start:
			switch ch {
			case CR:
				state = sw_almost_header_done
			case LF:
				goto header_done
			default:
				state = sw_header_name
			}
		case sw_header_name:
			switch ch {
			case ':':
				name = buf.Read(i)
				i = 0
				state = sw_space_before_value
			case LF:
				state = sw_almost_req_done
				i = 0
			}
		case sw_space_before_value:
			switch ch {
			case ' ':
				buf.Consume(2)
				i = 0
			default:
				state = sw_header_value
			}
		case sw_header_value:
			switch ch {
			case CR:
				header[string(name)] = buf.Read(i)
				//log.Printf("header name %s value %s", string(name), string(header[string(name)]))
				i = 0
				state = sw_almost_header_done
			}
		case sw_almost_header_done:
			if ch == LF {
				buf.Consume(2)
				i = 0
				state = sw_header_name
			}
		case sw_almost_req_done:
			switch ch {
			case LF:
				buf.Consume(2)
				goto header_done
			case CR:
			default:

				//log.Printf("before default")
				return Message{}, invalidReq
			}
		}
	}
	return Message{}, noData
header_done:
	h.header = header
	return Message{}, nil
}

type HttpUrl struct {
	schema   []byte
	host     []byte
	path     []byte
	query    []byte
	fragment []byte
	userinfo []byte
	params   map[string]string
}

func (h *HttpProtocol) ParseUrl(url []byte) (HttpUrl, error) {
	var hu HttpUrl
	n := bytes.Index(url, []byte("//"))
	if n < 0 {
		return hu, errors.New("bad schema")
	}
	hu.schema = url[:n]
	if bytes.IndexByte(hu.schema, '/') >= 0 {
		return hu, errors.New("bad schema")
	}
	if len(hu.schema) > 0 && hu.schema[len(hu.schema)-1] == ':' {
		hu.schema = hu.schema[:len(hu.schema)-1]
	}

	n += 2

	url = url[n:]

	n = bytes.IndexByte(url, '/')
	if n < 0 {

	}
	return hu, nil
}
