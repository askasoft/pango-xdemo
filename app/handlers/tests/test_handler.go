package tests

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/oss/mem"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

func Index(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "tests/index", h)
}

func Crash(c *xin.Context) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		if c.PostForm("crash") != "no" {
			panic("crash")
		}
		wg.Done()
	}()
	wg.Wait()

	c.String(http.StatusOK, "OK\n")
}

func Panic(c *xin.Context) {
	panic("panic")
}

func OutOfMemory(c *xin.Context) {
	mms, _ := mem.GetMemoryStats()

	limit := (mms.Total + mms.SwapTotal) * 1024
	alloc := 1024 * 1024 * 1024 // 1GB

	var rms runtime.MemStats

	var total uint64
	mm := map[int]string{}
	for i := 0; total < limit; i++ {
		mm[i] = str.Repeat(num.Itoa(i%10), alloc)
		total += uint64(alloc)

		runtime.ReadMemStats(&rms)
		c.Logger.Infof("malloc(%s) -> %s, heap: %s", num.HumanSize(alloc), num.HumanSize(total), num.HumanSize(rms.HeapAlloc))
		log.Flush()
	}

	c.String(http.StatusOK, fmt.Sprintf("malloc: %s", num.HumanSize(total)))
}

func StackOverflow(c *xin.Context) {
	var level uint64
	var fa func()
	var fb func()

	cnt := func() {
		level++
		if level%100000 == 0 {
			c.Logger.Infof("stack(%s)", num.Comma(level))
			log.Flush()
		}
	}

	fa = func() {
		cnt()
		fb()
	}
	fb = func() {
		cnt()
		fa()
	}

	fa()

	c.String(http.StatusOK, fmt.Sprintf("stack: %d", level))
}
