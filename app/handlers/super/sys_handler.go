package super

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

func SysStats(c *xin.Context) {
	h := handlers.H(c)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	sb := &str.Builder{}

	fmt.Fprintf(sb, "         CPU: %d\n", runtime.NumCPU())
	fmt.Fprintf(sb, "   Goroutine: %d\n", runtime.NumGoroutine())
	fmt.Fprintf(sb, "     Mallocs: %s\n", num.HumanSize(float64(m.Mallocs)))
	fmt.Fprintf(sb, "       Frees: %s\n", num.Comma(m.Frees))
	fmt.Fprintf(sb, "       Alloc: %s\n", num.HumanSize(float64(m.Alloc)))
	fmt.Fprintf(sb, "   HeapAlloc: %s\n", num.HumanSize(float64(m.HeapAlloc)))
	fmt.Fprintf(sb, "  TotalAlloc: %s\n", num.HumanSize(float64(m.TotalAlloc)))
	fmt.Fprintf(sb, "         Sys: %s\n", num.HumanSize(float64(m.Sys)))
	fmt.Fprintf(sb, "     Lookups: %s\n", num.Comma(m.Lookups))
	fmt.Fprintf(sb, "   HeapAlloc: %s\n", num.HumanSize(float64(m.HeapAlloc)))
	fmt.Fprintf(sb, "    HeapIdle: %s\n", num.HumanSize(float64(m.HeapIdle)))
	fmt.Fprintf(sb, "   HeapInuse: %s\n", num.HumanSize(float64(m.HeapInuse)))
	fmt.Fprintf(sb, "HeapReleased: %s\n", num.HumanSize(float64(m.HeapReleased)))
	fmt.Fprintf(sb, " HeapObjects: %s\n", num.Comma(m.HeapObjects))
	fmt.Fprintf(sb, "  StackInuse: %s\n", num.HumanSize(float64(m.StackInuse)))
	fmt.Fprintf(sb, "    StackSys: %s\n", num.HumanSize(float64(m.StackSys)))
	fmt.Fprintf(sb, "  MSpanInuse: %s\n", num.HumanSize(float64(m.MSpanInuse)))
	fmt.Fprintf(sb, "    MSpanSys: %s\n", num.HumanSize(float64(m.MSpanSys)))
	fmt.Fprintf(sb, " MCacheInuse: %s\n", num.HumanSize(float64(m.MCacheInuse)))
	fmt.Fprintf(sb, "   MCacheSys: %s\n", num.HumanSize(float64(m.MCacheSys)))

	h["Stats"] = sb.String()

	c.HTML(http.StatusOK, "super/sys", h)
}
