package tenant

import (
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/gwp"
)

var muWORKS sync.Mutex

func (tt *Tenant) GetWorkerPool() *gwp.WorkerPool {
	if wp, ok := app.WORKS.Get(string(tt.Schema)); ok {
		return wp
	}

	muWORKS.Lock()
	defer muWORKS.Unlock()

	// get again to prevent duplicated load
	if wp, ok := app.WORKS.Get(string(tt.Schema)); ok {
		return wp
	}

	wp := gwp.NewWorkerPool(tt.MaxWorkers(), 0)

	app.WORKS.Set(string(tt.Schema), wp)
	return wp
}
