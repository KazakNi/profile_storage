package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Server struct {
		Host string `yaml:"host" env:"SRV_HOST,HOST" env-description:"Server host" env-default:"localhost"`
		Port string `yaml:"port" env:"SRV_PORT,PORT" env-description:"Server port" env-default:"8080"`
	} `yaml:"server"`
	Token struct {
		Secret string `yaml:"secret"`
		Salt   string `yaml:"salt"`
	} `yaml:"token"`
	Swagger struct {
		HtmlPath   string `yaml:"htmlPath" env:"htmlPath" env-description:"Path to swagger html" env-default:"../internal/static/redoc.html"`
		StaticPath string `yaml:"staticPath" env:"staticPath" env-description:"Path to static folder" env-default:"../internal/static/"`
	}
}

var Cfg AppConfig

func LoadConfig() {

	err := cleanenv.ReadConfig("../config/config.yaml", &Cfg)
	if err != nil {
		fmt.Println(err)
		panic("Can't load config data")
	}
}
