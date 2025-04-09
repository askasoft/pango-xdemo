package super

import (
	"cmp"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/bol"
	"github.com/askasoft/pango/cas"
	"github.com/askasoft/pango/cog/treemap"
	"github.com/askasoft/pango/gwp"
	"github.com/askasoft/pango/imc"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tmu"
	"github.com/askasoft/pango/xin"
)

func ApstatsIndex(c *xin.Context) {
	h := handlers.H(c)
	h["JobStats"] = jobs.Stats()
	h["Caches"] = []string{"configs", "schemas", "workers", "users", "afips"}
	c.HTML(http.StatusOK, "super/apstats", h)
}

func ApstatsJobs(c *xin.Context) {
	c.String(http.StatusOK, jobs.Stats())
}

type CacheItem struct {
	Key string `json:"key,omitempty"`
	Val string `json:"val,omitempty"`
	TTL string `json:"ttl,omitempty"`
}

func apCacheStats[K comparable, T any](c *xin.Context, ic *imc.Cache[K, T], fv func(T) string) {
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

	c.JSON(http.StatusOK, xin.H{"size": cis.Len(), "data": cis})
}

func ApstatsConfigs(c *xin.Context) {
	apCacheStats(c, app.CONFS, func(v map[string]string) string {
		return num.Itoa(len(v))
	})
}

func ApstatsSchemas(c *xin.Context) {
	apCacheStats(c, app.SCMAS, bol.Btoa)
}

func ApstatsWorkers(c *xin.Context) {
	apCacheStats(c, app.WORKS, func(v *gwp.WorkerPool) string {
		return num.Itoa(v.CurWorks()) + "/" + num.Itoa(v.MaxWorks())
	})
}

func ApstatsUsers(c *xin.Context) {
	apCacheStats(c, app.USERS, func(v *models.User) string {
		return v.Role + ": " + v.Name
	})
}

func ApstatsAfips(c *xin.Context) {
	apCacheStats(c, app.AFIPS, func(v int) string {
		return num.Comma(v)
	})
}
