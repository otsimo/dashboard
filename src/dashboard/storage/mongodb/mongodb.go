package mongodb

import (
	"dashboard/storage"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MongoDBDriverName string = "mongodb"
	mongoURLFlag string = "mongodb-url"
)

func init() {
	storage.Register(MongoDBDriverName, &storage.RegisteredDriver{
		New: newMongoDriver,
		Flags: []cli.Flag{
			cli.StringFlag{Name: mongoURLFlag, Value: "mongodb://localhost:27017/Otsimo", Usage: "MongoDB url", EnvVar: "MONGODB_URL"},
		},
	})
}

func newMongoDriver(ctx *cli.Context) (storage.Driver, error) {
	url := ctx.String(mongoURLFlag)

	s, err := mgo.Dial(url)

	if err != nil {
		return nil, err
	}
	log.Debug("mongodb.go: connected to mongodb")
	md := &MongoDBDriver{
		Session: s,
	}
	return md, nil
}

type MongoDBDriver struct {
	Session *mgo.Session
}

func (d MongoDBDriver) Name() string {
	return MongoDBDriverName
}

func (d *MongoDBDriver) GetUser(id string) (*storage.DashboardUser, error) {
	c := d.Session.DB("").C("DashboardUser")
	u := &storage.DashboardUser{}
	if err := c.FindId(id).One(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (d *MongoDBDriver) GetItems(userID, provider string, from int64) ([]*storage.Item, error) {
	c := d.Session.DB("").C("DashboardItems")
	var res []*storage.Item
	fromTime := time.Unix(from, 0)
	if err := c.Find(bson.M{"user_id": userID, "provider_name": provider, "created_at": bson.M{"$gte": fromTime}}).All(res); err != nil {
		return []*storage.Item{}, err
	}
	return res, nil
}
