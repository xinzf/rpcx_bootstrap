package bootstrap

import (
	"errors"
	"gopkg.in/mgo.v2"
	"log"
)

var Mongo *mongo

type mongo struct {
	session *mgo.Session
}

type mongoConfig struct {
	Addr  string `mapstructure:"addr"`
	Debug bool   `mapstructure:"debug"`
}

func (this *mongo) Init() error {
	Mongo = new(mongo)
	if Config.Mongo.Addr == "" {
		return errors.New("mongo config's addr is empty")
	}

	//if Config.Mongo.Name == "" {
	//    return errors.New("mongo config's name is empty")
	//}

	var err error
	Mongo.session, err = mgo.Dial(Config.Mongo.Addr)
	if err != nil {
		return err
	}

	mgo.SetDebug(Config.Mongo.Debug)
	//mgo.SetLogger(log.New(os.Stderr,"mgo: ",log.LstdFlags))

	Mongo.session.SetMode(mgo.Monotonic, true)
	log.Println("Mongodb init success")
	return nil
}

//func (this *mongo) Use() *mgo.Database {
//    s := this.session.Copy()
//    return s.DB(Config.Mongo.Name)
//}

func (this *mongo) Close() {
	this.session.Close()
}
