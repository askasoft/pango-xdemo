package super

import (
	"cmp"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/askasoft/pango/bol"
	"github.com/askasoft/pango/cas"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/cog/treemap"
	"github.com/askasoft/pango/gwp"
	"github.com/askasoft/pango/imc"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/oss/cpu"
	"github.com/askasoft/pango/oss/disk"
	"github.com/askasoft/pango/oss/loadavg"
	"github.com/askasoft/pango/oss/mem"
	"github.com/askasoft/pango/oss/osm"
	"github.com/askasoft/pango/oss/uptime"
	"github.com/askasoft/pango/tmu"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox/xin"
)

func StatsIndex(c *xin.Context) {
	h := handlers.H(c)

	h["Server"] = statsServer()
	h["Caches"] = []string{"configs", "schemas", "workers", "users", "afips"}

	c.HTML(http.StatusOK, "super/stats", h)
}

func StatsServer(c *xin.Context) {
	c.JSON(http.StatusOK, statsServer())
}

func StatsJobs(c *xin.Context) {
	c.String(http.StatusOK, jobs.Stats())
}

func StatsDB(c *xin.Context) {
	c.JSON(http.StatusOK, app.SDB.Stats())
}

func statsServer() any {
	stm := linkedhashmap.NewLinkedHashMap[string, string]()

	// server
	host, _ := os.Hostname()
	stm.Set("Hostname", host)
	stm.Set("OS", runtime.GOOS)
	stm.Set("Arch", runtime.GOARCH)
	stm.Set("CPU", num.Comma(runtime.NumCPU()))

	var val string

	// uptime
	upt, err := uptime.GetUptime()
	if err != nil {
		val = err.Error()
	} else {
		val = tmu.HumanDuration(upt)
	}
	stm.Set("Uptime", val)

	// memory stats
	ms, err := mem.GetMemoryStats()
	if err != nil {
		val = err.Error()
	} else {
		val = fmt.Sprintf("%s / %s (%s%%)",
			num.HumanSize(ms.Used()),
			num.HumanSize(ms.Total),
			num.FtoaWithDigits(ms.Usage()*100, 2),
		)
	}
	stm.Set("Memory", val)

	// disk usage
	du, err := disk.GetDiskUsage(".")
	if err != nil {
		val = err.Error()
	} else {
		val = fmt.Sprintf("%s / %s (%s%%)",
			num.HumanSize(du.Used()),
			num.HumanSize(du.Total),
			num.FtoaWithDigits(du.Usage()*100, 2),
		)
	}
	stm.Set("Disk", val)

	// loadavg
	la, err := loadavg.GetLoadAvg()
	if err != nil {
		val = err.Error()
	} else {
		val = fmt.Sprintf("%.3f/1m, %.3f/5m, %.3f/15m", la.Loadavg1, la.Loadavg5, la.Loadavg15)
	}
	stm.Set("Load Average", val)

	// cpu stats
	if osm.Monitoring() {
		cu := osm.LastCPUUsage()
		val = fmt.Sprintf(
			"%.2f us, %.2f sy, %.2f ni, %.2f id, %.2f wa, %.2f hi, %.2f si, %.2f st, %.2f gu, %.2f gn",
			cu.UserUsage()*100, cu.SystemUsage()*100, cu.NiceUsage()*100, cu.IdleUsage()*100, cu.IowaitUsage()*100,
			cu.IrqUsage()*100, cu.SoftirqUsage()*100, cu.StealUsage()*100, cu.GuestUsage()*100, cu.GuestNiceUsage()*100,
		)
	} else {
		cu, err := cpu.GetCPUUsage(time.Millisecond * 250)
		if err != nil {
			val = err.Error()
		} else {
			val = fmt.Sprintf(
				"%.2f us, %.2f sy, %.2f ni, %.2f id, %.2f wa, %.2f hi, %.2f si, %.2f st, %.2f gu, %.2f gn",
				cu.UserUsage()*100, cu.SystemUsage()*100, cu.NiceUsage()*100, cu.IdleUsage()*100, cu.IowaitUsage()*100,
				cu.IrqUsage()*100, cu.SoftirqUsage()*100, cu.StealUsage()*100, cu.GuestUsage()*100, cu.GuestNiceUsage()*100,
			)
		}
	}
	stm.Set("%Cpu(s)", val)

	// network
	ifaces, err := net.Interfaces()
	if err == nil {
		var sb strings.Builder
		for _, i := range ifaces {
			if i.Flags&net.FlagLoopback != 0 || i.Flags&net.FlagRunning == 0 {
				continue
			}

			addrs, err := i.Addrs()
			if err != nil {
				continue
			}

			sb.Reset()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					sb.WriteString(addr.String())
					sb.WriteRune('\n')
				}
			}
			if sb.Len() > 0 {
				name := "Network #" + num.Itoa(i.Index) + " " + i.Name
				stm.Set(name, sb.String())
			}
		}
	}

	return stm
}

func StatsCacheConfigs(c *xin.Context) {
	statsCacheStats(c, app.CONFS, func(v map[string]string) string {
		return num.Itoa(len(v))
	})
}

func StatsCacheSchemas(c *xin.Context) {
	statsCacheStats(c, app.SCMAS, bol.Btoa)
}

func StatsCacheWorkers(c *xin.Context) {
	statsCacheStats(c, app.WORKS, func(v *gwp.WorkerPool) string {
		return num.Itoa(v.CurWorks()) + "/" + num.Itoa(v.MaxWorks())
	})
}

func StatsCacheUsers(c *xin.Context) {
	statsCacheStats(c, app.USERS, func(v *models.User) string {
		return v.Role + ": " + v.Name
	})
}

func StatsCacheAfips(c *xin.Context) {
	statsCacheStats(c, app.AFIPS, func(v int) string {
		return num.Comma(v)
	})
}

type CacheItem struct {
	Key string `json:"key,omitempty"`
	Val string `json:"val,omitempty"`
	TTL string `json:"ttl,omitempty"`
}

func statsCacheStats[K comparable, T any](c *xin.Context, ic *imc.Cache[K, T], fv func(T) string) {
	limit := num.Atoi(c.Query("limit"), 1000)

	now := time.Now()
	cis := treemap.NewTreeMap[string, CacheItem](cmp.Compare[string])

	var ci CacheItem
	ic.Each(func(k K, i imc.Item[T]) bool {
		s, _ := cas.ToString(k)
		ci.Key = s
		ci.Val = fv(i.Val)
		ci.TTL = tmu.HumanDuration(time.Unix(i.TTL, 0).Sub(now))
		cis.Set(s, ci)
		return cis.Len() <= limit
	})

	c.JSON(http.StatusOK, xin.H{"total": ic.Len(), "data": cis})
}
