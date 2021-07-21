package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type GbeConfig struct {
	DataSource DataSourceConfig `json:"dataSource"`
	Redis      RedisConfig      `json:"redis"`
	Kafka      KafkaConfig      `json:"kafka"`
	PushServer PushServerConfig `json:"pushServer"`
	RestServer RestServerConfig `json:"restServer"`
	JwtSecret  string           `json:"jwtSecret"`
	ApiKey     string           `json:"apiKey"`
}

type DataSourceConfig struct {
	DriverName        string `json:"driverName"`
	Addr              string `json:"addr"`
	Database          string `json:"database"`
	User              string `json:"user"`
	Password          string `json:"password"`
	EnableAutoMigrate bool   `json:"enableAutoMigrate"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
}

type PushServerConfig struct {
	Addr string `json:"addr"`
	Path string `json:"path"`
}

type RestServerConfig struct {
	Addr string `json:"addr"`
}

var config GbeConfig
var configOnce sync.Once

func GetConfig() *GbeConfig {
	configOnce.Do(func() {
		gopath := os.Getenv("GOPATH")
		bytes, err := ioutil.ReadFile(gopath + "/trade-engine/conf.json")
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			panic(err)
		}
	})
	return &config
}
