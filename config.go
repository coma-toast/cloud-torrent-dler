package main

import (
	"os"

	"github.com/spf13/viper"
)

// config is the configuration struct
type config struct {
	BaseURL         string
	DlRoot          string
	PidFilePath     string
	CachePath       string
	CompletedFolder string
	ShowRSS         string
	Username        string
	Passwd          string
	UseFTP          bool
	UseAPI          bool
}

func (c config) GetSeedrInstance() SeedrInstance {
	var selectedSeedr SeedrInstance
	if c.UseAPI {
		selectedSeedr = &SeedrAPI{
			Username: c.Username,
			Password: c.Passwd,
		}

		return selectedSeedr
	} else if c.UseFTP {
		selectedSeedr = &SeedrFTP{
			Username: c.Username,
			Password: c.Passwd,
		}
		return selectedSeedr
	}
	panic("Please provide a method")
}

// new config instance
var (
	conf *config
)

func getConf() *config {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()

	if err != nil {
		panic(err)
	}

	conf := &config{}
	err = viper.Unmarshal(conf)

	if err != nil {
		panic(err)
	}

	if conf.UseAPI && conf.UseFTP {
		panic("Cannot use both API and FTP. Pick one.")
	}

	if _, err := os.Stat(conf.DlRoot); os.IsNotExist(err) {
		os.Mkdir(conf.DlRoot, 0777)
	}

	return conf
}
