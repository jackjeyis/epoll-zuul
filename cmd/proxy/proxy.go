package main

import (
	"epoll-zuul/io"
	"fmt"
	"log"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
)

//_ "net/http/pprof"

var (
	pt = fmt.Printf
	ce = func(err error) {
		if err != nil {
			panic(err)
		}
	}
)

// set on build time
var (
	GitCommit = ""
	BuildTime = ""
	GoVersion = ""
	Version   = ""
)

// PrintVersion 输出版本信息
func PrintVersion() bool {
	fmt.Println("Version  : ", Version)
	fmt.Println("GitCommit: ", GitCommit)
	fmt.Println("BuildTime: ", BuildTime)
	fmt.Println("GoVersion: ", GoVersion)
	return true
}

func main() {

	start := func() {
		runtime.LockOSThread()

		sock, err := syscall.Socket(
			syscall.AF_INET,
			syscall.SOCK_STREAM,
			0,
		)
		ce(err)
		ce(syscall.SetsockoptInt(sock, syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1))
		ce(syscall.Bind(sock, &syscall.SockaddrInet4{
			Port: 8888,
			Addr: [4]byte{0, 0, 0, 0},
		}))
		ce(syscall.Listen(sock, 65536))
		defer syscall.Close(sock)

		epoll, err := syscall.EpollCreate1(0)
		ce(err)
		defer syscall.Close(epoll)

		ce(syscall.EpollCtl(epoll, syscall.EPOLL_CTL_ADD, sock, &syscall.EpollEvent{
			Events: syscall.EPOLLIN,
			Fd:     int32(sock),
		}))

		events := make([]syscall.EpollEvent, 1024)
		var fdLimit syscall.Rlimit
		ce(syscall.Getrlimit(syscall.RLIMIT_NOFILE, &fdLimit))
		pt("fdlimit %d\n", fdLimit)
		conns := make([]*io.Channel, fdLimit.Max)
		for {
			n, err := syscall.EpollWait(epoll, events, -1)
			ce(err)

			//start := time.Now().UnixNano()
			for _, ev := range events[:n] {
				//pt("epoll events fd %d,flag %d\n", ev.Fd, ev.Events)

				if ev.Events&syscall.EPOLLIN > 0 {
					if ev.Fd == int32(sock) {
						fd, _, err := syscall.Accept(sock)
						if err != nil {
							pt("accept: %v\n", err)
						}

						err = syscall.SetNonblock(fd, true)
						if err != nil {
							log.Printf("syscall.SetNonblock err %v", err)
						}
						err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
						if err != nil {
							log.Printf("syscall set nodelay err %v", err)
						}
						e := &syscall.EpollEvent{
							Events: syscall.EPOLLIN | syscall.EPOLLHUP | syscall.EPOLLRDHUP | syscall.EPOLLERR | -syscall.EPOLLET,
							Fd:     int32(fd),
						}
						ce(syscall.EpollCtl(epoll, syscall.EPOLL_CTL_ADD, fd, e))
						conns[e.Fd] = io.NewChannel(e)

					} else {
						if conn := conns[ev.Fd]; conn != nil {
							/*n, err := syscall.Read(int(ev.Fd), conn.Buf)
							//pt("read: %v\n,n bytes %d\n", err, n)
							if err != nil {
								pt("read: %v\n", err)
								//continue
							}
							conn.Feed(conn.Buf[:n])
							*/
							conn.OnRead()
						}

					}

				}

				if ev.Events&(syscall.EPOLLERR|syscall.EPOLLHUP|syscall.EPOLLRDHUP) > 0 {
					if conn := conns[ev.Fd]; conn != nil {
						ce(syscall.EpollCtl(epoll, syscall.EPOLL_CTL_DEL, int(ev.Fd), conn.Ev))
						conns[ev.Fd] = nil
						//bufPool.Put(conn.Buf)
					}
					syscall.Close(int(ev.Fd))
				}

			}
			//pt("time elapse %d", time.Now().UnixNano()-start)
		}
	}

	pt("cpm num %d \n", runtime.NumCPU())
	for i := 0; i < 1; i++ {
		go start()
	}

	select {}

}

/*
type Conn struct {
	Ev  *syscall.EpollEvent
	Buf []byte
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

func NewConn(ev *syscall.EpollEvent) *Conn {
	return &Conn{
		Ev:  ev,
		Buf: bufPool.Get().([]byte),
	}
}

func (c *Conn) Feed(data []byte) {
	if _, err := syscall.Write(int(c.Ev.Fd), data); err != nil {
		pt("write: %v\n", err)
	}
}
*/
