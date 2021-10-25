package settings

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var Cfg Config

type Config struct {
	RabbitMQ struct {
		User        string `yaml:"user"`
		Pass        string `yaml:"pass"`
		Host        string `yaml:"host"`
		Port        string `yaml:"port"`
		ClientQueue string `yaml:"clientQueue"`
		StooqQueue  string `yaml:"stooqQueue"`
	}
	Database struct {
		User       string `yaml:"user"`
		Pass       string `yaml:"pass"`
		Protocol   string `yaml:"protocol"`
		Host       string `yaml:"host"`
		Port       string `yaml:"port"`
		DataSource string `yaml:"dataSource"`
	}
	Server struct {
		SecretKey string `yaml:"secretKey"`
	}
}

func GetConfig() *Config {
	var cfg Config
	readFile(&cfg)

	Cfg = cfg

	return &cfg
}

func readFile(cfg *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
