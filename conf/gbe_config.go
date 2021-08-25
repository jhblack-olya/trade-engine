package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"flag"
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

		configFile := flag.String("config", "", "run with config file, refer to README.md file")
		flag.Parse()
		var fileName string
		if configFile != nil {
			fileName = *configFile
		}
		if flag.NFlag() != 1 {
			fileName = "/conf.json"
		}

		pwd, _ := os.Getwd()
		fmt.Println("Loaded file: ", fileName)
		bytes, err := ioutil.ReadFile(pwd + fileName)
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
