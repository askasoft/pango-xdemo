package schema

import (
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox/xfs"
	"github.com/askasoft/pangox/xjm"
)

var tables = linkedhashmap.NewLinkedHashMap[string, any]()

func init() {
	tables.Set("files", &xfs.File{})
	tables.Set("jobs", &xjm.Job{})
	tables.Set("job_logs", &xjm.JobLog{})
	tables.Set("job_chains", &xjm.JobChain{})
	tables.Set("users", &models.User{})
	tables.Set("configs", &models.Config{})
	tables.Set("audit_logs", &models.AuditLog{})
	tables.Set("pets", &models.Pet{})
}

func (sm Schema) Prefix() string {
	if len(sm) == 0 {
		return ""
	}
	return string(sm) + "."
}

func (sm Schema) Table(s string) string {
	return sm.Prefix() + s
}

func (sm Schema) TableFiles() string {
	return sm.Table("files")
}

func (sm Schema) TableJobs() string {
	return sm.Table("jobs")
}

func (sm Schema) TableJobLogs() string {
	return sm.Table("job_logs")
}

func (sm Schema) TableJobChains() string {
	return sm.Table("job_chains")
}

func (sm Schema) TableUsers() string {
	return sm.Table("users")
}

func (sm Schema) TableConfigs() string {
	return sm.Table("configs")
}

func (sm Schema) TableAuditLogs() string {
	return sm.Table("audit_logs")
}

func (sm Schema) TablePets() string {
	return sm.Table("pets")
}
