package bootstrap

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"reflect"
)

type Cfg struct {
	Server   ServerConfig   `mapstructure:"server"`
	Register registerConfig `mapstructure:"register"`
	Db       *dbConfig      `mapstructure:"db"`
	Redis    *redisConfig   `mapstructure:"redis"`
	Mongo    *mongoConfig   `mapstructure:"mongo"`
	Client   clientConfig   `mapstructure:"client"`
	Logger   loggerConfig   `mapstructure:"logger"`
	Mode     string         `mapstrucure:"mod"`
}

type ServerConfig struct {
	Name         string `mapstructure:"name"`
	Proto        string `mapstructure:"proto"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type registerConfig struct {
	Addr string `mapstructure:"addr"`
	//TimeOut int64 `mapstructure:"addr"`
}

func (s ServerConfig) String() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func ReadConfig(fileName string, dstConf interface{}, initFn ...func(interface{}) error) error {
	viper.SetConfigName(fileName)
	viper.AddConfigPath(*confFilePath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	refType := reflect.TypeOf(dstConf)
	if refType.Kind() != reflect.Ptr {
		return errors.New("config's argument: dstConf is not a pointer")
	}

	if err := viper.Unmarshal(dstConf); err != nil {
		return err
	}

	for _, fn := range initFn {
		if err := fn(dstConf); err != nil {
			return err
		}
	}

	return nil
}
