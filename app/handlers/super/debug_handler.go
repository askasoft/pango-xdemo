package super

import (
	"fmt"
	"net/http"
	hprof "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/oss/cpu"
	"github.com/askasoft/pango/oss/disk"
	"github.com/askasoft/pango/oss/loadavg"
	"github.com/askasoft/pango/oss/mem"
	"github.com/askasoft/pango/oss/uptime"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tmu"
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

func debugSystem() any {
	// system
	host, _ := os.Hostname()
	stm := linkedhashmap.NewLinkedHashMap[string, string]()
	stm.Set("Host", host)
	stm.Set("OS", runtime.GOOS)
	stm.Set("Arch", runtime.GOARCH)
	stm.Set("CPU", num.Comma(runtime.NumCPU()))

	var val string

	// memory usage
	mu, err := mem.GetMemoryUsage()
	if err != nil {
		val = err.Error()
	} else {
		val = num.HumanSize(mu.Free()) + " / " + num.HumanSize(mu.Total())
	}
	stm.Set("Memory (Free / Total)", val)

	// disk usage
	du, err := disk.GetDiskUsage(".")
	if err != nil {
		val = err.Error()
	} else {
		val = num.HumanSize(du.Used()) + " / " + num.HumanSize(du.Available()) + " / " + num.HumanSize(du.Total()) + " / " + num.FtoaWithDigits(du.Usage()*100, 2) + "%"
	}
	stm.Set("Disk (Used / Available / Total / Usage)", val)

	return stm
}

func debugRuntime() any {
	// runtime
	rtm := linkedhashmap.NewLinkedHashMap[string, string]()
	rtm.Set("Goversion", runtime.Version())
	rtm.Set("Gomaxprocs", num.Comma(runtime.GOMAXPROCS(0)))
	rtm.Set("Goroutine", num.Comma(runtime.NumGoroutine()))
	rtm.Set("Cmdline", str.Join(os.Args, " "))
	rtm.Set("Startup", fmt.Sprintf("%s (%s)", app.StartupTime.Format(time.RFC3339), tmu.HumanDuration(time.Since(app.StartupTime))))

	upt, err := uptime.GetUptime()
	if err != nil {
		rtm.Set("Uptime", err.Error())
	} else {
		rtm.Set("Uptime", tmu.HumanDuration(upt))
	}

	return rtm
}

func debugCPUStats() any {
	csm := linkedhashmap.NewLinkedHashMap[string, string]()

	var val string

	// cpu stats
	cs, err := cpu.GetCPUStats()
	if err != nil {
		val = err.Error()
	} else {
		val = fmt.Sprintf(
			"%.2f us, %.2f sy, %.2f ni, %.2f id, %.2f wa, %.2f hi, %.2f si, %.2f st, %.2f%%",
			cs.UserUsage(), cs.SystemUsage(), cs.NiceUsage(), cs.IdleUsage(), cs.IowaitUsage(),
			cs.IrqUsage(), cs.SoftirqUsage(), cs.StealUsage(), cs.CPUUsage(),
		)
	}
	csm.Set("%Cpu(s)", val)

	// loadavg
	la, err := loadavg.GetLoadAvg()
	if err != nil {
		val = err.Error()
	} else {
		val = fmt.Sprintf("%.3f / %.3f / %.3f", la.Loadavg1, la.Loadavg5, la.Loadavg15)
	}
	csm.Set("Load Average (1 / 5 / 15)", val)

	return csm
}

func debugMemStats() any {
	// memory
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	msm := linkedhashmap.NewLinkedHashMap[string, string]()
	msm.Set("Mallocs", num.HumanSize(float64(ms.Mallocs)))
	msm.Set("Frees", num.Comma(ms.Frees))
	msm.Set("Alloc", num.HumanSize(float64(ms.Alloc)))
	msm.Set("HeapAlloc", num.HumanSize(float64(ms.HeapAlloc)))
	msm.Set("TotalAlloc", num.HumanSize(float64(ms.TotalAlloc)))
	msm.Set("Sys", num.HumanSize(float64(ms.Sys)))
	msm.Set("Lookups", num.Comma(ms.Lookups))
	msm.Set("HeapAlloc", num.HumanSize(float64(ms.HeapAlloc)))
	msm.Set("HeapIdle", num.HumanSize(float64(ms.HeapIdle)))
	msm.Set("HeapInuse", num.HumanSize(float64(ms.HeapInuse)))
	msm.Set("HeapReleased", num.HumanSize(float64(ms.HeapReleased)))
	msm.Set("HeapObjects", num.Comma(ms.HeapObjects))
	msm.Set("StackInuse", num.HumanSize(float64(ms.StackInuse)))
	msm.Set("StackSys", num.HumanSize(float64(ms.StackSys)))
	msm.Set("MSpanInuse", num.HumanSize(float64(ms.MSpanInuse)))
	msm.Set("MSpanSys", num.HumanSize(float64(ms.MSpanSys)))
	msm.Set("MCacheInuse", num.HumanSize(float64(ms.MCacheInuse)))
	msm.Set("MCacheSys", num.HumanSize(float64(ms.MCacheSys)))
	msm.Set("GCSys", num.HumanSize(ms.GCSys))
	msm.Set("LastGC", time.Unix(0, int64(ms.LastGC)).Format(time.RFC3339Nano)) //nolint: gosec
	msm.Set("NextGC", num.HumanSize(ms.NextGC))
	msm.Set("NumGC", num.Comma(ms.NumGC))
	msm.Set("GCPauseTotal", tmu.HumanDuration(time.Duration(ms.PauseTotalNs))) //nolint: gosec
	if sec := time.Since(app.StartupTime).Seconds(); sec != 0 {
		msm.Set("GCNumPerSecond", num.Comma(float64(ms.NumGC)/sec, 6))
		msm.Set("GCPausePerSecond", tmu.HumanDuration(time.Duration(ms.PauseTotalNs/uint64(sec)))) //nolint: gosec
	}
	return msm
}

func debugProfiles() any {
	// profiles
	var pfs []*profile
	for _, p := range pprof.Profiles() {
		pfs = append(pfs, &profile{p.Name(), profileDescriptions[p.Name()], p.Count()})
	}
	sort.Slice(pfs, func(i, j int) bool {
		return pfs[i].Name < pfs[j].Name
	})
	return pfs
}

func debugTrace() any {
	// traces
	trs := []*trace{
		{"profile", traceDescriptions["profile"]},
		{"symbol", traceDescriptions["symbol"]},
		{"trace", traceDescriptions["trace"]},
	}
	return trs
}

func DebugIndex(c *xin.Context) {
	h := handlers.H(c)

	h["System"] = debugSystem()
	h["Runtime"] = debugRuntime()
	h["CPUStats"] = debugCPUStats()
	h["MemStats"] = debugMemStats()
	h["Profiles"] = debugProfiles()
	h["Traces"] = debugTrace()

	c.HTML(http.StatusOK, "super/debug", h)
}

func DebugJSON(c *xin.Context) {
	stats := make(map[string]any)

	stats["system"] = debugSystem()
	stats["runtime"] = debugRuntime()
	stats["cpustats"] = debugCPUStats()
	stats["memstats"] = debugMemStats()

	c.JSON(http.StatusOK, stats)
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
