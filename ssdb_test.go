package ssgo

import (
	"math/rand"
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

func TestClientMultiFunc(t *testing.T) {
	cn, _ := pool.GetClient()
	defer cn.Release()

	cn.MultiHSet("test1", ss1{
		X:  11,
		I:  22,
		U:  33,
		S:  "ss1",
		Bt: true,
		P:  []byte("sss3"),
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

func TestBatchDo(t *testing.T) {
	cn, _ := pool.GetClient()
	defer cn.Release()
	batch := BatchExec{
		{"set", "test1", "aabb"},
		{"get", "test1"},
		{"del", "test1"},
		{"get", "test1"},
	}
	reps, e := cn.BatchDo(batch)
	if e != nil {
		t.Log(e)
	}
	for _, r := range reps {
		t.Logf("%s : %v\n", r.R.String(), r.E)
	}
}

func makeValue(size int) []byte {
	val := make([]byte, size)
	for i := 0; i < size; i++ {
		str := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
		val[i] = str[rand.Int31n(int32(62))]
	}
	return val
}

func TestBigValue(t *testing.T) {
	v := makeValue(1024 * 1024)
	cn, _ := pool.GetClient()
	defer cn.Release()
	batch := BatchExec{
		{"set", "testBig", v},
		{"get", "testBig"},
		{"del", "testBig"},
	}
	reps, e := cn.BatchDo(batch)
	if e != nil {
		t.Log(e)
	}
	if reps[1].R.String() != string(v) {
		t.Error("get value error")
	}
}

var (
	ssgoTestKey1 = "ssgo.test.key1"
	ssgoTestKey2 = "ssgo.test.key2"
)

func benchmarkSSDBGo(valSize, batchSize int, b *testing.B) {
	val := makeValue(valSize)
	cn, _ := pool.GetClient()
	defer cn.Release()
	defer cn.Do("hclear", ssgoTestKey1)
	defer cn.Do("hclear", ssgoTestKey2)
	b.Logf("insert %d recorde", b.N*batchSize)
	for n := 0; n < b.N*batchSize; n++ {
		cn.Do("hset", ssgoTestKey1, n, val)
	}
	//	return
	b.ResetTimer()
	lastKey := ""
	for n := 0; n < b.N; n++ {
		rep, e := cn.Do("hscan", ssgoTestKey1, lastKey, "", batchSize)
		if e != nil {
			b.Error(e)
		}
		if rep == nil {
			b.Error("rep nil")
		}
		entrys := rep.Hash()
		if len(entrys) == 0 {
			continue
		}
		for _, ent := range entrys {
			cn.Do("hset", ssgoTestKey2, ent.Key, ent.Value)
			cn.Do("hdel", ssgoTestKey1, ent.Key)
		}
	}

}

// go test -bench=.
// go test -bench=BenchmarkSSDBGo_1k_10 -benchtime=20s
// go test -run=none -bench=BenchmarkSSDBGo_10k_50 -cpuprofile=cprof
func BenchmarkSSDBGo_1k_10(b *testing.B) {
	benchmarkSSDBGo(1*1024, 10, b)
}
func BenchmarkSSDBGo_1k_50(b *testing.B) {
	benchmarkSSDBGo(1*1024, 50, b)
}

func BenchmarkSSDBGo_10k_10(b *testing.B) {
	benchmarkSSDBGo(10*1024, 10, b)
}

func BenchmarkSSDBGo_10k_50(b *testing.B) {
	benchmarkSSDBGo(10*1024, 50, b)
}
