package ssgo

import (
	"bufio"
	"net"
	"runtime"
	"time"
)

type ConPool struct {
	cType    string
	cAddr    string
	cTimeout time.Duration
	conns    chan *Client
}

func NewConPool(hostAddr string, maxConn int) *ConPool {

	if maxConn < 1 {
		maxConn = runtime.NumCPU() * 2
	}

	cr := &ConPool{
		cType:    "tcp",
		cAddr:    hostAddr,
		cTimeout: time.Duration(30) * time.Second,
		conns:    make(chan *Client, maxConn),
	}

	return cr
}

func dialTimeout(network, addr string) (*Client, error) {

	rAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}
	sock, err := net.DialTCP(network, nil, rAddr)
	if err != nil {
		return nil, err
	}

	return &Client{sock: sock, reader: bufio.NewReader(sock)}, nil
}

func (cr *ConPool) dialNew() (*Client, error) {
	cn, err := dialTimeout(cr.cType, cr.cAddr)
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

	return cn.Do(args...)
}

func (cr *ConPool) BatchDo(batch BatchExec) ([]ReplyE, error) {
	cn, e := cr.GetClient()
	if e != nil {
		return nil, e
	}
	defer cn.Release()

	return cn.BatchDo(batch)
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
