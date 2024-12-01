package super

import (
	"net/http"
	"runtime"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/xin"
)

func SysStats(c *xin.Context) {
	h := handlers.H(c)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	lmap := linkedhashmap.NewLinkedHashMap[string, string]()

	lmap.Set("CPU", num.Comma(runtime.NumCPU()))
	lmap.Set("Goroutine", num.Comma(runtime.NumGoroutine()))
	lmap.Set("Mallocs", num.HumanSize(float64(m.Mallocs)))
	lmap.Set("Frees", num.Comma(m.Frees))
	lmap.Set("Alloc", num.HumanSize(float64(m.Alloc)))
	lmap.Set("HeapAlloc", num.HumanSize(float64(m.HeapAlloc)))
	lmap.Set("TotalAlloc", num.HumanSize(float64(m.TotalAlloc)))
	lmap.Set("Sys", num.HumanSize(float64(m.Sys)))
	lmap.Set("Lookups", num.Comma(m.Lookups))
	lmap.Set("HeapAlloc", num.HumanSize(float64(m.HeapAlloc)))
	lmap.Set("HeapIdle", num.HumanSize(float64(m.HeapIdle)))
	lmap.Set("HeapInuse", num.HumanSize(float64(m.HeapInuse)))
	lmap.Set("HeapReleased", num.HumanSize(float64(m.HeapReleased)))
	lmap.Set("HeapObjects", num.Comma(m.HeapObjects))
	lmap.Set("StackInuse", num.HumanSize(float64(m.StackInuse)))
	lmap.Set("StackSys", num.HumanSize(float64(m.StackSys)))
	lmap.Set("MSpanInuse", num.HumanSize(float64(m.MSpanInuse)))
	lmap.Set("MSpanSys", num.HumanSize(float64(m.MSpanSys)))
	lmap.Set("MCacheInuse", num.HumanSize(float64(m.MCacheInuse)))
	lmap.Set("MCacheSys", num.HumanSize(float64(m.MCacheSys)))

	h["Stats"] = lmap

	c.HTML(http.StatusOK, "super/sys", h)
}
