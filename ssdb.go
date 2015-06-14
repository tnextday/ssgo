package ssgo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

var (
	ErrProtocolError = errors.New("ssdb protocol error")
)

type Client struct {
	reader   *bufio.Reader
	sock     *net.TCPConn
	recv_buf bytes.Buffer
	pool     *ConPool
	netErr   error
}

type BatchExec [][]interface{}

func Connect(ip string, port int) (*Client, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}
	sock, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	var c Client
	c.sock = sock
	c.reader = bufio.NewReader(sock)
	return &c, nil
}

func (c *Client) Do(args ...interface{}) (Reply, error) {

	if err := c.send(args); err != nil {
		return nil, err
	}

	resp, err := c.recv()
	if err != nil || len(resp) < 1 {
		return nil, err
	}
	if resp[0] != "ok" {
		return nil, errors.New(resp[0])
	}
	return resp[1:], nil
}

func (c *Client) DoIgnoreErr(args ...interface{}) Reply {
	r, e := c.Do(args...)
	if e != nil {
		return Reply{}
	}
	return r
}

func (c *Client) BatchDo(batch BatchExec) (reps []ReplyE, e error) {
	l := len(batch)
	replys := make([]ReplyE, l)
	errCount := 0

	for i, args := range batch {
		replys[i].R, replys[i].E = c.Do(args...)
		if replys[i].E != nil {
			errCount++
		}
	}

	if errCount != 0 {
		e = fmt.Errorf("BatchDo: get %d errors", errCount)
	}
	return replys, e
}

func (c *Client) Set(key string, val string) (interface{}, error) {
	resp, err := c.Do("set", key, val)
	if err != nil {
		return nil, err
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return true, nil
	}
	return nil, fmt.Errorf("bad response")
}

// TODO: Will somebody write addition semantic methods?
func (c *Client) Get(key string) (interface{}, error) {
	resp, err := c.Do("get", key)
	if err != nil {
		return nil, err
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1], nil
	}
	if resp[0] == "not_found" {
		return nil, nil
	}
	return nil, fmt.Errorf("bad response")
}

func (c *Client) Del(key string) (interface{}, error) {
	resp, err := c.Do("del", key)
	if err != nil {
		return nil, err
	}

	//response looks like this: [ok 1]
	if len(resp) > 0 && resp[0] == "ok" {
		return true, nil
	}
	return nil, fmt.Errorf("bad response:resp:%v:", resp)
}

func (c *Client) MultiHSet(name string, obj interface{}, keys ...string) error {
	args := Args{"multi_hset", name}.AddFlat(obj, keys...)
	_, e := c.Do(args...)
	return e
}

func (c *Client) MultiHGet(name string, obj interface{}, keys ...string) error {
	var (
		rep Reply
		e   error
	)
	if len(keys) > 0 {
		args := []interface{}{"multi_hget", name}
		for _, v := range keys {
			args = append(args, v)
		}
		rep, e = c.Do(args...)
	} else {
		rep, e = c.Do("hgetall", name)
	}
	if e != nil {
		return e
	}
	return ScanStruct(rep, obj)
}

func (c *Client) GenAutoIncId(name string) int64 {
	rep, e := c.Do("incr", name)
	if e != nil {
		return 0
	}
	return rep.Int64()
}

func (c *Client) Send(args ...interface{}) error {
	return c.send(args)
}

func (c *Client) send(args []interface{}) error {
	var buf bytes.Buffer
	for _, arg := range args {
		var s string
		switch arg := arg.(type) {
		case string:
			s = arg
		case []byte:
			s = string(arg)
		case []string:
			for _, s := range arg {
				buf.WriteString(fmt.Sprintf("%d", len(s)))
				buf.WriteByte('\n')
				buf.WriteString(s)
				buf.WriteByte('\n')
			}
			continue
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			s = fmt.Sprintf("%d", arg)
		case float32, float64, complex64, complex128:
			s = fmt.Sprintf("%f", arg)
		case bool:
			if arg {
				s = "1"
			} else {
				s = "0"
			}
		case nil:
			s = ""
		default:
			buf, e := json.Marshal(arg)
			if e != nil {
				return fmt.Errorf("bad arguments")
			}
			s = string(buf)
		}
		buf.WriteString(fmt.Sprintf("%d", len(s)))
		buf.WriteByte('\n')
		buf.WriteString(s)
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')
	_, err := c.sock.Write(buf.Bytes())
	if err != nil {
		c.netErr = err
	}
	return err
}

func (c *Client) Recv() ([]string, error) {
	return c.recv()
}

func (c *Client) recv() ([]string, error) {
	resp := []string{}
	bb := bytes.NewBuffer(nil)
	for {
		l, _, e := c.reader.ReadLine()
		if e != nil {
			return nil, e
		}
		if len(l) == 0 {
			//empty line found
			break
		}
		size, e := strconv.Atoi(string(l))
		if e != nil {
			return nil, e
		}
		if size < 0 {
			return nil, ErrProtocolError
		}
		bb.Reset()
		_, e = io.CopyN(bb, c.reader, int64(size+1))
		if e != nil {
			return nil, e
		}
		buf := bb.Bytes()
		if buf[size] != '\n' {
			return nil, ErrProtocolError
		}
//		fmt.Println("read buf:", string(bb.Bytes()[:size]))

		resp = append(resp, string(buf[:size]))
	}
	return resp, nil
}


// Close The Client Connection
func (c *Client) close() error {
	return c.sock.Close()
}

func (c *Client) Release() error {
	if c.netErr != nil {
		// if client have net error, try to close it
		return c.close()
	} else if c.pool != nil {
		c.pool.push(c)
		return nil
	}
	return c.close()
}

func MakeKey(args ...interface{}) string {
	ss := make([]string, 0, len(args))
	for _, arg := range args {
		var s string
		switch arg := arg.(type) {
		case string:
			s = arg
		case []byte:
			s = string(arg)
		case []string:
			s = strings.Join(arg, "_")
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			s = fmt.Sprintf("%d", arg)
		case float32, float64, complex64, complex128:
			s = fmt.Sprintf("%f", arg)
		case bool:
			if arg {
				s = "1"
			} else {
				s = "0"
			}
			//		case nil:
			//			s = ""
		default:
			//TODO how to do ?
			//			s = ""
		}
		if s != "" {

			ss = append(ss, strings.TrimSpace(s))
		}

	}
	return strings.Join(ss, ":")
}
