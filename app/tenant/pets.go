package tenant

import (
	"math/rand"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"gorm.io/gorm"
)

func (tt Tenant) ResetPets(logger log.Logger) error {
	err := app.GDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Transaction(func(db *gorm.DB) error {
		gfs := tt.GFS(db)
		logger.Infof("Delete Pet Files: /%s", models.PrefixPetFile)
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

		if err := initPets(logger, db, 1000, "dog"); err != nil {
			return err
		}

		if err := initPets(logger, db, 2000, "cat"); err != nil {
			return err
		}

		return nil
	})

	return err
}

func initPets(logger log.Logger, db *gorm.DB, cid int64, cat string) error {
	// ipath := "./pets/"

	// imgs := []string{}
	// for i := 1; ; i++ {
	// 	fn := filepath.Join(ipath, cat+str.PadLeft(num.Itoa(i), 2, "0")+".jpg")
	// 	if err := fsu.FileExists(fn); err != nil {
	// 		break
	// 	}
	// 	imgs = append(imgs, fn)
	// }

	pgs := tbsutil.GetPetGenderMap("").Keys()
	pos := tbsutil.GetPetOriginMap("").Keys()
	pts := tbsutil.GetPetTemperMap("").Keys()
	phs := tbsutil.GetPetHabitsMap("").Keys()

	bd, _ := time.Parse(time.RFC3339, "2000-01-01T10:04:05+09:00")
	for i := 0; i < 100; i++ {
		//		File f = files.get(Randoms.randInt(files.size()));

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
		}

		logger.Infof("Create Pet: #%d %s", pet.ID, pet.Name)
		if err := db.Create(pet).Error; err != nil {
			return err
		}
		// // Pet Image
		// PetImage pi = new PetImage();
		// pi.setId(p.getId());
		// pi.setPid(p.getId());
		// pi.setImageName(f.getName());
		// pi.setImageSize((int)f.length());
		// pi.setImageData(Files.readToBytes(f));
		// assist().setCreatedByFields(pi);

		// dao.insert(pi);
		// status.count++;
		// printInfo("Add PetImage: " + pi.getId() + " / " + pi.getImageName());
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
