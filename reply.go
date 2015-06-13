package ssgo

import (
	"encoding/json"
	"strconv"
)

//const (
//	ReplyOK          string = "ok"
//	ReplyNotFound    string = "not_found"
//	ReplyError       string = "error"
//	ReplyFail        string = "fail"
//	ReplyClientError string = "client_error"
//)

type Reply []string

type ReplyE struct {
	R Reply
	E error
}

type Entry struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (r Reply) String() string {

	if len(r) > 0 {
		return r[0]
	}

	return ""
}

func (r Reply) Int() int {
	return int(r.Int64())
}

func (r Reply) Int64() int64 {

	if len(r) < 1 {
		return 0
	}

	i64, err := strconv.ParseInt(r[0], 10, 64)
	if err == nil {
		return i64
	}

	return 0
}

func (r Reply) Uint() uint {
	return uint(r.Uint64())
}

func (r Reply) Uint64() uint64 {

	if len(r) < 1 {
		return 0
	}

	i64, err := strconv.ParseUint(r[0], 10, 64)
	if err == nil {
		return i64
	}

	return 0
}

func (r Reply) Float64() float64 {

	if len(r) < 1 {
		return 0
	}

	f64, err := strconv.ParseFloat(r[0], 64)
	if err == nil {
		return f64
	}

	return 0
}

func (r Reply) Bool() bool {

	if len(r) < 1 {
		return false
	}

	b, err := strconv.ParseBool(r[0])
	if err == nil {
		return b
	}

	return false
}

func (r Reply) List() []string {

	if len(r) < 1 {
		return []string{}
	}

	return r
}

func (r Reply) Hash() []*Entry {

	l := len(r)

	if l < 2 {
		return []*Entry{}
	}
	hs := make([]*Entry, 0, l / 2)

	for i := 0; i < (l - 1); i += 2 {
		hs = append(hs, &Entry{r[i], r[i+1]})
	}

	return hs
}

func (r Reply) Map() map[string]string {

	m := map[string]string{}

	if len(r) < 2 {
		return m
	}

	for i := 0; i < (len(r) - 1); i += 2 {
		m[r[i]] = r[i+1]
	}

	return m
}

// Json returns the map that marshals from the reply bytes as json in response .
func (r Reply) Json(v interface{}) error {
	return json.Unmarshal([]byte(r.String()), &v)
}

func (r *Entry) Json(v interface{}) error {
	return json.Unmarshal([]byte(r.Value), &v)
}
