package args

import (
	"time"

	"github.com/askasoft/pango/cog/hashset"
	"github.com/askasoft/pango/doc/jsonx"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
)

type UserQueryArg struct {
	QueryArg

	ID       string   `json:"id,omitempty" form:"id,strip,ascii" validate:"uintegers"`
	Role     []string `json:"role,omitempty" form:"role,strip"`
	Status   []string `json:"status,omitempty" form:"status,strip"`
	LoginMFA []string `json:"login_mfa,omitempty" form:"login_mfa"`
	CIDR     string   `json:"cidr,omitempty" form:"cidr,strip"`
	Name     string   `json:"name,omitempty" form:"name,strip"`
	Email    string   `json:"email,omitempty" form:"email,strip"`
}

func (uqa *UserQueryArg) String() string {
	return jsonx.Stringify(uqa)
}

func (uqa *UserQueryArg) HasFilters() bool {
	return uqa.ID != "" ||
		len(uqa.Role) > 0 ||
		len(uqa.Status) > 0 ||
		len(uqa.LoginMFA) > 0 ||
		uqa.CIDR != "" ||
		uqa.Name != "" ||
		uqa.Email != ""
}

func (uqa *UserQueryArg) AddFilters(sqb *sqlx.Builder) {
	uqa.AddIntegers(sqb, "id", uqa.ID)
	uqa.AddIn(sqb, "role", uqa.Role)
	uqa.AddIn(sqb, "status", uqa.Status)
	uqa.AddIn(sqb, "login_mfa", uqa.LoginMFA)
	uqa.AddKeywords(sqb, "cidr", uqa.CIDR)
	uqa.AddKeywords(sqb, "name", uqa.Name)
	uqa.AddKeywords(sqb, "email", uqa.Email)
}

type AuditLogQueryArg struct {
	QueryArg

	DateFrom time.Time `json:"date_from,omitempty" form:"date_from,strip"`
	DateTo   time.Time `json:"date_to,omitempty" form:"date_to,strip"`
	Func     []string  `json:"func,omitempty" form:"func,strip"`
	Action   string    `json:"action,omitempty" form:"action,strip"`
	User     string    `json:"user,omitempty" form:"user,strip"`
	CIP      string    `json:"cip,omitempty" form:"cip,strip"`
}

func (alqa *AuditLogQueryArg) HasFilters() bool {
	return !alqa.DateFrom.IsZero() ||
		!alqa.DateTo.IsZero() ||
		len(alqa.Func) > 0 ||
		alqa.Action != "" ||
		alqa.User != "" ||
		alqa.CIP != ""
}

func (alqa *AuditLogQueryArg) AddFilters(sqb *sqlx.Builder, locale string) {
	alqa.AddDateRange(sqb, "audit_logs.date", alqa.DateFrom, alqa.DateTo)
	alqa.AddKeywords(sqb, "users.email", alqa.User)
	alqa.AddKeywords(sqb, "audit_logs.cip", alqa.CIP)
	alqa.AddIn(sqb, "audit_logs.func", alqa.Func)

	if alqa.Action != "" {
		ass := str.Fields(alqa.Action)
		acm := tbsutil.GetAudioLogFunactMap(locale)

		fam := map[string]*hashset.HashSet[string]{}
		for _, a := range ass {
			for k, v := range acm {
				if str.ContainsFold(v, a) {
					fun, act, _ := str.CutByte(k, '.')
					if fa, ok := fam[fun]; ok {
						fa.Add(act)
					} else {
						fam[fun] = hashset.NewHashSet(act)
					}
				}
			}
		}

		if len(fam) > 0 {
			var sb str.Builder
			var args []any

			sb.WriteByte('(')
			for fun, acts := range fam {
				if sb.Len() > 1 {
					sb.WriteString(" OR ")
				}

				sin, ains := sqx.In("audit_logs.action", acts.Values())
				sb.WriteString("(audit_logs.func = ? AND ")
				sb.WriteString(sin)
				sb.WriteString(")")
				args = append(args, fun)
				args = append(args, ains...)
			}
			sb.WriteByte(')')

			sqb.Where(sb.String(), args...)
		}
	}
}

type FileQueryArg struct {
	QueryArg

	ID       string    `json:"id,omitempty" form:"id,strip,ascii" validate:"uintegers"`
	Name     string    `json:"name,omitempty" form:"name,strip"`
	Ext      string    `json:"ext,omitempty" form:"ext,strip"`
	Size     string    `json:"size,omitempty" form:"size,strip,ascii" validate:"uintegers"`
	TimeFrom time.Time `json:"time_from,omitempty" form:"time_from,strip"`
	TimeTo   time.Time `json:"time_to,omitempty" form:"time_to,strip" validate:"omitempty,gtefield=TimeFrom"`
}

func (fqa *FileQueryArg) String() string {
	return jsonx.Stringify(fqa)
}

func (fqa *FileQueryArg) HasFilters() bool {
	return fqa.ID != "" ||
		fqa.Name != "" ||
		fqa.Ext != "" ||
		fqa.Size != "" ||
		!fqa.TimeFrom.IsZero() ||
		!fqa.TimeTo.IsZero()
}

func (fqa *FileQueryArg) AddFilters(sqb *sqlx.Builder) {
	fqa.AddKeywords(sqb, "id", fqa.ID)
	fqa.AddKeywords(sqb, "name", fqa.Name)
	fqa.AddKeywords(sqb, "ext", fqa.Ext)
	fqa.AddIntegers(sqb, "size", fqa.Size)
	fqa.AddTimeRange(sqb, "time", fqa.TimeFrom, fqa.TimeTo)
}

type PetQueryArg struct {
	QueryArg

	ID       string    `json:"id,omitempty" form:"id,strip,ascii" validate:"uintegers"`
	Name     string    `json:"name,omitempty" form:"name,strip"`
	BornFrom time.Time `json:"born_from,omitempty" form:"born_from,strip"`
	BornTo   time.Time `json:"born_to,omitempty" form:"born_to,strip" validate:"omitempty,gtefield=BornFrom"`
	Gender   []string  `json:"gender,omitempty" form:"gender,strip"`
	Origin   []string  `json:"origin,omitempty" form:"origin,strip"`
	Habits   []string  `json:"habits,omitempty" form:"habits,strip"`
	Temper   []string  `json:"temper,omitempty" form:"temper,strip"`
	Amount   string    `json:"amount,omitempty" form:"amount,strip,ascii" validate:"uintegers"`
	Price    string    `json:"price,omitempty" form:"price,strip,ascii" validate:"decimals"`
	ShopName string    `json:"shop_name,omitempty" form:"shop_name,strip"`
}

func (pqa *PetQueryArg) String() string {
	return jsonx.Stringify(pqa)
}

func (pqa *PetQueryArg) HasFilters() bool {
	return pqa.ID != "" ||
		pqa.Name != "" ||
		len(pqa.Gender) > 0 ||
		len(pqa.Origin) > 0 ||
		len(pqa.Temper) > 0 ||
		len(pqa.Habits) > 0 ||
		!pqa.BornFrom.IsZero() ||
		!pqa.BornTo.IsZero() ||
		pqa.Amount != "" ||
		pqa.Price != "" ||
		pqa.ShopName != ""
}

func (pqa *PetQueryArg) AddFilters(sqb *sqlx.Builder) {
	pqa.AddIntegers(sqb, "id", pqa.ID)
	pqa.AddIn(sqb, "gender", pqa.Gender)
	pqa.AddIn(sqb, "origin", pqa.Origin)
	pqa.AddIn(sqb, "temper", pqa.Temper)
	pqa.AddTimeRange(sqb, "born_at", pqa.BornFrom, pqa.BornTo)
	pqa.AddIntegers(sqb, "amount", pqa.Amount)
	pqa.AddDecimals(sqb, "price", pqa.Price)
	pqa.AddKeywords(sqb, "name", pqa.Name)
	pqa.AddKeywords(sqb, "shop_name", pqa.ShopName)
	pqa.AddContainsAll(sqb, "habits", pqa.Habits)
}
