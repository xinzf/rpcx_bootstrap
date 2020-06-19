package bootstrap

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"time"
)

type database struct {
	connection *gorm.DB
}

type dbConfig struct {
	Addr         string `mapstructure:"addr"`
	User         string `mapstructure:"user"`
	Pswd         string `mapstructure:"pswd"`
	Name         string `mapstructure:"name"`
	Log          bool   `mapstructure:"log"`
	MaxIdleConns int    `mapstructure:"maxidle_conns"`
	MaxOpenConns int    `mapstructure:"maxopen_conns"`
}

func (s *dbConfig) String() string {
	u := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&interpolateParams=true&parseTime=true&loc=Local",
		s.User,
		s.Pswd,
		s.Addr,
		s.Name)
	return u
}

var DB *database

func (db *database) Init() error {
	DB = &database{}

	if Config.Db.Addr == "" {
		return errors.New("dbconfig's addr is empty")
	}
	if Config.Db.User == "" {
		return errors.New("dbconfig's user is empty")
	}
	if Config.Db.Name == "" {
		return errors.New("dbconfig's name is empty")
	}
	if Config.Db.Pswd == "" {
		return errors.New("dbconfig's pswd is empty")
	}

	d, err := gorm.Open("mysql", Config.Db.String())
	if err != nil {
		return err
	}

	d.LogMode(Config.Db.Log)

	unixMilli := func(t time.Time) int64 {
		return t.UnixNano() / 1e6
	}

	d.DB().SetMaxIdleConns(Config.Db.MaxIdleConns)
	d.DB().SetMaxOpenConns(Config.Db.MaxOpenConns)
	d.DB().SetConnMaxLifetime(time.Duration(300) * time.Second)

	d.Callback().Create().Before("gorm:create").Register("set_created_updated", func(scope *gorm.Scope) {
		if scope.HasColumn("created") {
			scope.SetColumn("created", unixMilli(time.Now()))
		}
		if scope.HasColumn("updated") {
			scope.SetColumn("updated", unixMilli(time.Now()))
		}
	})
	d.Callback().Update().Before("gorm:update").Register("set_updated", func(scope *gorm.Scope) {
		if scope.HasColumn("updated") {
			scope.SetColumn("updated", unixMilli(time.Now()))
		}
	})

	d.SingularTable(true)

	DB.connection = d
	return nil
}

func (db *database) Use() *gorm.DB {
	return db.connection
}

func (db *database) Close() {
	DB.connection.Close()
}
