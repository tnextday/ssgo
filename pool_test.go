package ssgo
import (
	"testing"
)

func pingBenchmark(t *testing.T, poolCache, parallel, times int)  {
	t.Logf("Ping test %d, %d, %d\n", poolCache, parallel, times)
	pool := NewConPool("127.0.0.1:8888", poolCache)
	sem := make(chan int, parallel)
	for i := 0; i < times; i++{
		sem <- 1
		go func() {
			cn, err := pool.GetClient()
			if err != nil {
				t.Error(err)
			}
			defer cn.Release()
			_, e := cn.Do("ping")
			if e != nil{
				t.Error(e)
			}
			<-sem
		}()
	}
	for i := 0; i < parallel; i++{
		sem <- 1
	}
	pool.Close()
}

func TestConnPool(t *testing.T) {
	pingBenchmark(t, 3, 2, 1000)
	pingBenchmark(t, 3, 10, 1000)
	pingBenchmark(t, 3, 100, 1000)
}

