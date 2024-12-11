package super

import (
	"net/http"
	hprof "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/sys/du"
	"github.com/askasoft/pango/sys/mu"
	"github.com/askasoft/pango/xin"
)

var profileDescriptions = map[string]string{
	"allocs":       "A sampling of all past memory allocations",
	"block":        "Stack traces that led to blocking on synchronization primitives",
	"goroutine":    "Stack traces of all current goroutines. Use debug=2 as a query parameter to export in the same format as an unrecovered panic.",
	"heap":         "A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.",
	"mutex":        "Stack traces of holders of contended mutexes",
	"threadcreate": "Stack traces that led to the creation of new OS threads",
}

var traceDescriptions = map[string]string{
	"profile": "CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.",
	"symbol":  "Looks up the program counters listed in the request, responding with a table mapping program counters to function names.",
	"trace":   "A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.",
}

type profile struct {
	Name  string
	Desc  string
	Count int
}

type trace struct {
	Name string
	Desc string
}

func DebugIndex(c *xin.Context) {
	h := handlers.H(c)

	// system
	host, _ := os.Hostname()
	stm := linkedhashmap.NewLinkedHashMap[string, string]()
	stm.Set("Host", host)
	stm.Set("CPU", num.Comma(runtime.NumCPU()))
	stm.Set("Memory (Free / Total)", num.HumanSize(mu.FreeMemory())+" / "+num.HumanSize(mu.TotalMemory()))

	// disk usage
	du := du.NewDiskUsage(".")
	stm.Set(
		"Disk (Used / Available / Total / Usage)",
		num.HumanSize(du.Used())+" / "+num.HumanSize(du.Available())+" / "+num.HumanSize(du.Total())+" / "+num.FtoaWithDigits(du.Usage()*100, 2)+"%",
	)

	// runtime
	rtm := linkedhashmap.NewLinkedHashMap[string, string]()
	rtm.Set("Goroutine", num.Comma(runtime.NumGoroutine()))
	rtm.Set("Cmdline", str.Join(os.Args, " "))

	// memory
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

	// profiles
	var pfs []*profile
	for _, p := range pprof.Profiles() {
		pfs = append(pfs, &profile{p.Name(), profileDescriptions[p.Name()], p.Count()})
	}
	sort.Slice(pfs, func(i, j int) bool {
		return pfs[i].Name < pfs[j].Name
	})

	// traces
	trs := []*trace{
		{"profile", traceDescriptions["profile"]},
		{"symbol", traceDescriptions["symbol"]},
		{"trace", traceDescriptions["trace"]},
	}

	h["System"] = stm
	h["Runtime"] = rtm
	h["Profiles"] = pfs
	h["Traces"] = trs
	h["MemStats"] = msm

	c.HTML(http.StatusOK, "super/debug", h)
}

func DebugPprof(c *xin.Context) {
	w, r := c.Writer, c.Request

	p := c.Param("prof")
	switch p {
	case "profile":
		hprof.Profile(w, r)
	case "symbol":
		hprof.Symbol(w, r)
	case "trace":
		hprof.Trace(w, r)
	case "cmdline":
		hprof.Cmdline(w, r)
	default:
		hprof.Handler(p).ServeHTTP(w, r)
	}
}
