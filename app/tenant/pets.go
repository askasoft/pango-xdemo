package tenant

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
	"gorm.io/gorm"
)

func (tt Tenant) ResetPets(logger log.Logger) error {
	err := app.GDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Transaction(func(db *gorm.DB) error {
		logger.Infof("Delete Pet Files: /%s", models.PrefixPetFile)
		gfs := tt.GFS(db)
		if _, err := gfs.DeletePrefix("/" + models.PrefixPetFile + "/"); err != nil {
			return err
		}

		logger.Info("Delete Pets")
		if err := db.Exec("TRUNCATE TABLE " + tt.TablePets()).Error; err != nil {
			return err
		}

		logger.Infof("Reset Pets Sequence")
		if err := db.Exec(tt.ResetSequence("pets")).Error; err != nil {
			return err
		}

		if err := tt.initPets(logger, db, 1000, "dog"); err != nil {
			return err
		}

		if err := tt.initPets(logger, db, 2000, "cat"); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (tt Tenant) initPets(logger log.Logger, db *gorm.DB, cid int64, cat string) error {
	ipath := "./data/pets/"

	imgs := []string{}
	for i := 1; ; i++ {
		fn := filepath.Join(ipath, fmt.Sprintf("%s%02d.jpg", cat, i))
		if err := fsu.FileExists(fn); err != nil {
			break
		}
		imgs = append(imgs, fn)
	}

	gfs := tt.GFS(db)

	pgs := tbsutil.GetPetGenderMap("").Keys()
	pos := tbsutil.GetPetOriginMap("").Keys()
	pts := tbsutil.GetPetTemperMap("").Keys()
	phs := tbsutil.GetPetHabitsMap("").Keys()

	bd, _ := time.Parse(time.RFC3339, "2000-01-01T10:04:05+09:00")
	for i := 0; i < 100; i++ {
		pet := &models.Pet{
			ID:          cid*1000 + int64(i),
			Name:        cat + " " + str.PadLeft(num.Itoa(i), 2, "0") + " " + petRandText(5),
			Gender:      pgs[rand.Intn(len(pgs))], //nolint: gosec
			BornAt:      bd.AddDate(0, 0, 1),
			Origin:      pos[rand.Intn(len(pos))],                                                            //nolint: gosec
			Temper:      pts[rand.Intn(len(pts))],                                                            //nolint: gosec
			Habits:      cog.NewHashSet[string](phs[rand.Intn(len(phs))], phs[rand.Intn(len(phs))]).Values(), //nolint: gosec
			Amount:      rand.Intn(100),                                                                      //nolint: gosec
			Price:       rand.Float64() * 10000,                                                              //nolint: gosec
			ShopName:    petRandText(10),
			Description: petRandText(64),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		logger.Infof("Create Pet: #%d %s", pet.ID, pet.Name)
		if err := db.Create(pet).Error; err != nil {
			return err
		}

		if len(imgs) > 0 {
			img := imgs[rand.Intn(len(imgs))] //nolint: gosec
			if _, err := xfs.SaveLocalFile(gfs, pet.PhotoPath(), img); err != nil {
				return err
			}
		}
	}

	return nil
}

func petRandText(n int) string {
	pns := "日月火水木金土赤青黄紫黒白藍天地村香川河海湖洋左右宇宙羽雨峰影用意容易花果中華快楽"
	cnt := str.RuneCount(pns)

	sb := &strings.Builder{}
	for i := 0; i < n; i++ {
		x := rand.Intn(cnt) //nolint: gosec
		sb.WriteString(str.Mid(pns, x, 1))
	}

	return str.Strip(sb.String())
}
