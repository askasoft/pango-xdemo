package args

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/utils/strutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog/hashset"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

type UserQueryArg struct {
	QueryArg

	ID       string   `json:"id,omitempty" form:"id,strip"`
	Name     string   `json:"name,omitempty" form:"name,strip"`
	Email    string   `json:"email,omitempty" form:"email,strip"`
	Role     []string `json:"role,omitempty" form:"role,strip"`
	Status   []string `json:"status,omitempty" form:"status,strip"`
	LoginMFA []string `json:"login_mfa,omitempty" form:"login_mfa"`
	CIDR     string   `json:"cidr,omitempty" form:"cidr,strip"`
}

func (uqa *UserQueryArg) String() string {
	return strutil.JSONString(uqa)
}

func (uqa *UserQueryArg) HasFilters() bool {
	return uqa.ID != "" ||
		uqa.Name != "" ||
		uqa.Email != "" ||
		len(uqa.Role) > 0 ||
		len(uqa.Status) > 0 ||
		len(uqa.LoginMFA) > 0 ||
		uqa.CIDR != ""
}

func (uqa *UserQueryArg) AddFilters(sqb *sqlx.Builder) {
	uqa.AddIDs(sqb, "id", uqa.ID)
	uqa.AddLikes(sqb, "name", uqa.Name)
	uqa.AddLikes(sqb, "email", uqa.Email)
	uqa.AddIn(sqb, "role", uqa.Role)
	uqa.AddIn(sqb, "status", uqa.Status)
	uqa.AddIn(sqb, "login_mfa", uqa.LoginMFA)
	uqa.AddLikes(sqb, "cidr", uqa.CIDR)
}

type AuditLogQueryArg struct {
	QueryArg

	ID       string    `json:"id,omitempty" form:"id,strip"`
	DateFrom time.Time `json:"date_from,omitempty" form:"date_from,strip"`
	DateTo   time.Time `json:"date_to,omitempty" form:"date_to,strip" validate:"omitempty,gtefield=DateFrom"`
	User     string    `json:"user,omitempty" form:"user,strip"`
	CIP      string    `json:"cip,omitempty" form:"cip,strip"`
	Func     []string  `json:"func,omitempty" form:"func,strip"`
	Action   string    `json:"action,omitempty" form:"action,strip"`
}

func (alqa *AuditLogQueryArg) HasFilters() bool {
	return alqa.ID != "" ||
		!alqa.DateFrom.IsZero() ||
		!alqa.DateTo.IsZero() ||
		alqa.User != "" ||
		alqa.CIP != "" ||
		len(alqa.Func) > 0 ||
		alqa.Action != ""
}

func (alqa *AuditLogQueryArg) AddFilters(sqb *sqlx.Builder, locale string) {
	alqa.AddIDs(sqb, "audit_logs.id", alqa.ID)
	alqa.AddDates(sqb, "audit_logs.date", alqa.DateFrom, alqa.DateTo)
	alqa.AddLikes(sqb, "users.email", alqa.User)
	alqa.AddLikes(sqb, "audit_logs.cip", alqa.CIP)
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

type PetQueryArg struct {
	QueryArg

	ID        string    `json:"id,omitempty" form:"id,strip"`
	Name      string    `json:"name,omitempty" form:"name,strip"`
	BornFrom  time.Time `json:"born_from,omitempty" form:"born_from,strip"`
	BornTo    time.Time `json:"born_to,omitempty" form:"born_to,strip" validate:"omitempty,gtefield=BornFrom"`
	Gender    []string  `json:"gender,omitempty" form:"gender,strip"`
	Origin    []string  `json:"origin,omitempty" form:"origin,strip"`
	Habits    []string  `json:"habits,omitempty" form:"habits,strip"`
	Temper    []string  `json:"temper,omitempty" form:"temper,strip"`
	AmountMin string    `json:"amount_min,omitempty" form:"amount_min"`
	AmountMax string    `json:"amount_max,omitempty" form:"amount_max"`
	PriceMin  string    `json:"price_min,omitempty" form:"price_min"`
	PriceMax  string    `json:"price_max,omitempty" form:"price_max"`
	ShopName  string    `json:"shop_name,omitempty" form:"shop_name,strip"`
}

func (pqa *PetQueryArg) String() string {
	return strutil.JSONString(pqa)
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
		pqa.AmountMin != "" ||
		pqa.AmountMax != "" ||
		pqa.PriceMin != "" ||
		pqa.PriceMax != "" ||
		pqa.ShopName != ""
}

func (pqa *PetQueryArg) AddFilters(sqb *sqlx.Builder) {
	pqa.AddIDs(sqb, "id", pqa.ID)
	pqa.AddIn(sqb, "gender", pqa.Gender)
	pqa.AddIn(sqb, "origin", pqa.Origin)
	pqa.AddIn(sqb, "temper", pqa.Temper)
	pqa.AddTimes(sqb, "born_at", pqa.BornFrom, pqa.BornTo)
	pqa.AddInts(sqb, "amount", pqa.AmountMin, pqa.AmountMax)
	pqa.AddFloats(sqb, "price", pqa.PriceMin, pqa.PriceMax)
	pqa.AddLikes(sqb, "name", pqa.Name)
	pqa.AddLikes(sqb, "shop_name", pqa.ShopName)
	pqa.AddContains(sqb, "habits", pqa.Habits)
}
