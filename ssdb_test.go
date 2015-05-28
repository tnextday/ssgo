package ssgo
import (
	"testing"
)

var (
	pool = NewConPool("127.0.0.1:8888", 3)
)

type ss0 struct {
	X  int
	Y  int `ssgo:"y"`
	Bt bool
}

type ss1 struct {
	X  int    `json:"-"`
	I  int    `json:"i"`
	U  uint   `json:"u"`
	S  string `ssgo:"s"`
	P  []byte `json:"p"`
	B  bool   `json:"b"`
	Bt bool
	Bf bool
	ss0
}
func TestClient(t *testing.T) {
	cn, _ := pool.GetClient()
	defer cn.Release()

	cn.MultiHSet("test1", ss1{
		X:11,
		I:22,
		U:33,
		S:"ss1",
		Bt:true,
		P:[]byte("sss3"),
	})

	s := ss1{}
	e := cn.MultiHGet("test1", &s)
	if e != nil {
		t.Error(e)
	}
	t.Log("%v\n", s)

	s2 := ss1{}
	e = cn.MultiHGet("test1", &s2, "u", "s")
	if e != nil {
		t.Error(e)
	}
	t.Log("%v\n", s2)
}