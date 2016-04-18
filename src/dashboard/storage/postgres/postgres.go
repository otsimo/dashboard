package postgres

import (
	"dashboard/storage"
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jinzhu/gorm"
)

const (
	PostgresDriverName string = "postgres"
	dsnFlag            string = "postgres-dsn"
)

func init() {
	storage.Register(PostgresDriverName, &storage.RegisteredDriver{
		New: newPostgresDriver,
		Flags: []cli.Flag{
			cli.StringFlag{Name: dsnFlag, Value: "postgres://localhost:5432/postgres", Usage: "Postgres db dsn", EnvVar: "POSTGRES_DSN"},
		},
	})
}

func newPostgresDriver(ctx *cli.Context) (storage.Driver, error) {
	url := ctx.String(dsnFlag)

	db, err := gorm.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	it := storage.Item{}
	if !db.HasTable(&it) {
		db.CreateTable(&it)
	}

	logrus.Debug("postgres.go: connected to db")
	md := &PostgresDriver{
		db: db,
	}
	return md, nil
}

type PostgresDriver struct {
	db *gorm.DB
}

func (d PostgresDriver) Name() string {
	return PostgresDriverName
}

func (d *PostgresDriver) GetUser(id string) (*storage.DashboardUser, error) {
	return []*storage.Item{}, errors.New("Not Implemented")
}

func (d *PostgresDriver) GetItems(userID, provider string, from int64) ([]*storage.Item, error) {
	return []*storage.Item{}, errors.New("Not Implemented")
}
