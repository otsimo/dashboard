package postgres

import (
	"dashboard/storage"
	"fmt"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	DB     *gorm.DB
	driver storage.Driver
)

func init() {
	var err error
	if DB, err = OpenTestConnection(); err != nil {
		panic(fmt.Sprintf("No error should happen when connecting to test database, but got err=%+v", err))
	}
	var it storage.Item
	var du storage.DashboardUser
	var ui storage.ProviderUserInfo
	DB.DropTableIfExists(&it, &du, &ui)

	driver, err = postgresDriverFromDB(DB)
	if err != nil {
		panic(fmt.Sprintf("Failed to create db driver, and got err=%+v", err))
	}
	if os.Getenv("DEBUG") == "true" {
		DB.LogMode(true)
	}
	DB.DB().SetMaxIdleConns(10)
	u := storage.DashboardUser{ID: "1234", Providers: []storage.ProviderUserInfo{
		{
			UserID: "1234",
			Name:   "pro1",
		},
	}}
	if err := DB.Model(&u).Create(&u).Error; err != nil {
		panic(err)
	}
}

func OpenTestConnection() (db *gorm.DB, err error) {
	dia := os.Getenv("TEST_DIALECT")
	dsn := os.Getenv("TEST_DB_DSN")
	if dia == "" {
		dia = "postgres"
	}
	db, err = gorm.Open(dia, dsn)
	return
}

func TestGetUser(t *testing.T) {
	du := driver.GetUser("1234")
	if len(du.Providers) != 1 {
		t.Fatal("must have one provider")
	}
	if du.Providers[0].Name != "pro1" {
		t.Error("Wrong provider name")
	}
}
