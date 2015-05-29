package ssgo

import (
	"net"
	"runtime"
	"time"
)

type ConPool struct {
	ctype    string
	clink    string
	ctimeout time.Duration
	conns    chan *Client
}

func NewConPool(hostAddr string, maxConn int) *ConPool {

	if maxConn < 1 {
		maxConn = runtime.NumCPU() * 2
	}

	cr := &ConPool{
		ctype:    "tcp",
		clink:    hostAddr,
		ctimeout: time.Duration(10) * time.Second,
		conns:    make(chan *Client, maxConn),
	}

	//	if cr.ctimeout < 1*time.Second {
	//		cr.ctimeout = 10 * time.Second
	//	}

	return cr
}

func dialTimeout(network, addr string) (*Client, error) {

	raddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}
	sock, err := net.DialTCP(network, nil, raddr)
	if err != nil {
		return nil, err
	}

	return &Client{sock: sock}, nil
}

func (cr *ConPool) dialNew() (*Client, error) {
	cn, err := dialTimeout(cr.ctype, cr.clink)
	if err != nil {
		return nil, err
	}
	cn.pool = cr
	return cn, nil
}

func (cr *ConPool) Do(args ...interface{}) (Reply, error) {

	cn, e := cr.GetClient()
	if e != nil {
		return nil, e
	}
	defer cn.Release()

	cn.sock.SetReadDeadline(time.Now().Add(cr.ctimeout))
	cn.sock.SetWriteDeadline(time.Now().Add(cr.ctimeout))

	return cn.Do(args...)
}

func (cr *ConPool) Close() {
	var conn *Client
	for {
		select {
		case conn = <-cr.conns:
			conn.close()
		default:
			return
		}
	}
}

func (cr *ConPool) push(cn *Client) {
	select {
	case cr.conns <- cn:
	default:
		cn.close()
	}
}

func (cr *ConPool) GetClient() (cn *Client, err error) {
	select {
	case conn := <-cr.conns:
		return conn, nil
	default:
		return cr.dialNew()
	}
}
