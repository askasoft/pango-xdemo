package super

import (
	"net/http"
	"runtime"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/fsu/du"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/xin"
)

func SysStats(c *xin.Context) {
	h := handlers.H(c)

	rtm := linkedhashmap.NewLinkedHashMap[string, string]()
	rtm.Set("CPU", num.Comma(runtime.NumCPU()))
	rtm.Set("Goroutine", num.Comma(runtime.NumGoroutine()))

	du := du.NewDiskUsage(".")
	dum := linkedhashmap.NewLinkedHashMap[string, string]()
	dum.Set("Total", num.HumanSize(du.Total()))
	dum.Set("Available", num.HumanSize(du.Available()))
	dum.Set("Used", num.HumanSize(du.Used()))
	dum.Set("Usage", num.FtoaWithDigits(du.Usage()*100, 2)+"%")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	msm := linkedhashmap.NewLinkedHashMap[string, string]()
	msm.Set("Mallocs", num.HumanSize(float64(m.Mallocs)))
	msm.Set("Frees", num.Comma(m.Frees))
	msm.Set("Alloc", num.HumanSize(float64(m.Alloc)))
	msm.Set("HeapAlloc", num.HumanSize(float64(m.HeapAlloc)))
	msm.Set("TotalAlloc", num.HumanSize(float64(m.TotalAlloc)))
	msm.Set("Sys", num.HumanSize(float64(m.Sys)))
	msm.Set("Lookups", num.Comma(m.Lookups))
	msm.Set("HeapAlloc", num.HumanSize(float64(m.HeapAlloc)))
	msm.Set("HeapIdle", num.HumanSize(float64(m.HeapIdle)))
	msm.Set("HeapInuse", num.HumanSize(float64(m.HeapInuse)))
	msm.Set("HeapReleased", num.HumanSize(float64(m.HeapReleased)))
	msm.Set("HeapObjects", num.Comma(m.HeapObjects))
	msm.Set("StackInuse", num.HumanSize(float64(m.StackInuse)))
	msm.Set("StackSys", num.HumanSize(float64(m.StackSys)))
	msm.Set("MSpanInuse", num.HumanSize(float64(m.MSpanInuse)))
	msm.Set("MSpanSys", num.HumanSize(float64(m.MSpanSys)))
	msm.Set("MCacheInuse", num.HumanSize(float64(m.MCacheInuse)))
	msm.Set("MCacheSys", num.HumanSize(float64(m.MCacheSys)))

	h["Runtime"] = rtm
	h["DiskUsage"] = dum
	h["MemStats"] = msm

	c.HTML(http.StatusOK, "super/sys", h)
}
