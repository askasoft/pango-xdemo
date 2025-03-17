package jobs

import (
	"mime/multipart"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
)

type IArg interface {
	Bind(c *xin.Context) error
}

type JobArgCreater func(*tenant.Tenant) IArg

var jobArgCreators = map[string]JobArgCreater{}

func RegisterJobArg(name string, jac JobArgCreater) {
	jobArgCreators[name] = jac
}

type IArgFile interface {
	GetFile() string
}

type ArgFile struct {
	File string `json:"file,omitempty" form:"-"`
}

func (af *ArgFile) GetFile() string {
	return af.File
}

func (af *ArgFile) SetFile(tt *tenant.Tenant, mfh *multipart.FileHeader) error {
	fid := app.MakeFileID(models.PrefixJobFile, mfh.Filename)
	tfs := tt.FS()
	if _, err := xfs.SaveUploadedFile(tfs, fid, mfh); err != nil {
		return err
	}

	af.File = fid
	return nil
}

type ArgItems struct {
	Items int `json:"items,omitempty" form:"items,strip" validate:"min=0"`
}

type ArgIDRange struct {
	IdFrom int64 `json:"id_from,omitempty" form:"id_from,strip" validate:"min=0"`
	IdTo   int64 `json:"id_to,omitempty" form:"id_to,strip" validate:"omitempty,min=0,gtefield=IdFrom"`
}

type iPeriod interface {
	Period() *ArgPeriod
}

type ArgPeriod struct {
	Start time.Time `json:"start,omitempty" form:"start"`
	End   time.Time `json:"end,omitempty" form:"end" validate:"omitempty,gtefield=Start"`
}

func (ap *ArgPeriod) Period() *ArgPeriod {
	return ap
}

func ArgBind(c *xin.Context, a any) error {
	err := c.Bind(a)

	if ip, ok := a.(iPeriod); ok {
		ap := ip.Period()
		if !ap.End.IsZero() {
			ap.End = ap.End.Add(time.Hour*24 - time.Microsecond)
		}
	}

	return err
}
