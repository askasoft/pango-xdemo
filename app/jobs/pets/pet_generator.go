package pets

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog/hashset"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
)

type PetGenerator struct {
	tt *tenant.Tenant

	cat  string
	pgs  []string
	pos  []string
	pts  []string
	phs  []string
	imgs []string
}

func NewPetGenerator(tt *tenant.Tenant, cat string) *PetGenerator {
	pg := &PetGenerator{tt: tt, cat: cat}

	pg.pgs = tbsutil.GetPetGenderMap("").Keys()
	pg.pos = tbsutil.GetPetOriginMap("").Keys()
	pg.pts = tbsutil.GetPetTemperMap("").Keys()
	pg.phs = tbsutil.GetPetHabitsMap("").Keys()

	ipath := "./data/pets/"

	for i := 1; ; i++ {
		fn := filepath.Join(ipath, fmt.Sprintf("%s%02d.jpg", pg.cat, i))
		if err := fsu.FileExists(fn); err != nil {
			break
		}
		pg.imgs = append(pg.imgs, fn)
	}

	return pg
}

func (pg *PetGenerator) Create(logger log.Logger, db *sqlx.DB, js *jobs.JobState) error {
	sfs := pg.tt.SFS(db)

	bd, _ := time.Parse(time.RFC3339, "2000-01-01T10:04:05+09:00")
	pet := &models.Pet{
		Name:        pg.cat + " " + str.PadLeft(num.Itoa(js.Step), 2, "0") + " " + pg.randText(5),
		Gender:      pg.pgs[rand.Intn(len(pg.pgs))], //nolint: gosec
		BornAt:      bd.AddDate(0, 0, js.Step),
		Origin:      pg.pos[rand.Intn(len(pg.pos))],                                                              //nolint: gosec
		Temper:      pg.pts[rand.Intn(len(pg.pts))],                                                              //nolint: gosec
		Habits:      hashset.NewHashSet(pg.phs[rand.Intn(len(pg.phs))], pg.phs[rand.Intn(len(pg.phs))]).Values(), //nolint: gosec
		Amount:      rand.Intn(100),                                                                              //nolint: gosec
		Price:       rand.Float64() * 10000,                                                                      //nolint: gosec
		ShopName:    pg.randText(10),
		Description: pg.randText(64),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	js.Step++
	logger.Infof("%s Create Pet: %s", js.Progress(), pet.Name)

	sqb := db.Builder()
	sqb.Insert(pg.tt.TablePets())
	sqb.StructNames(pet, "id")
	if !db.SupportLastInsertID() {
		sqb.Returns("id")
	}
	sql := sqb.SQL()

	pid, err := db.NamedCreate(sql, pet)
	if err != nil {
		return err
	}

	pet.ID = pid

	js.IncSuccess()
	logger.Infof("%s Pet #%d Created: %s", js.Progress(), pet.ID, pet.Name)

	if len(pg.imgs) > 0 {
		img := pg.imgs[rand.Intn(len(pg.imgs))] //nolint: gosec
		if _, err := xfs.SaveLocalFile(sfs, pet.PhotoPath(), img); err != nil {
			return err
		}
	}

	return nil
}

func (pg *PetGenerator) randText(n int) string {
	pns := "日月火水木金土赤青黄紫黒白藍天地村香川河海湖洋左右宇宙羽雨峰影用意容易花果中華快楽"
	cnt := str.RuneCount(pns)

	sb := &strings.Builder{}
	for i := 0; i < n; i++ {
		x := rand.Intn(cnt) //nolint: gosec
		sb.WriteString(str.Mid(pns, x, 1))
	}

	return str.Strip(sb.String())
}
