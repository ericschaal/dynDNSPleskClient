package main

import (
	"os"
	"github.com/creamdog/gonfig"
)


type Config struct {
	Server string
	Delay int
	Port string
	Token string
}


func readConfig() Config {

	var config Config

	f, err := os.Open("/etc/dynDNSClient.json")
	if err != nil {
		errlog.Println("Couldn't open config file.")
		os.Exit(1);
	}
	defer f.Close();
	rawConfig, err := gonfig.FromJson(f)
	if err != nil {
		errlog.Println("Couldn't serialize configuration.")
	}
	if err := rawConfig.GetAs("", &config); err != nil {
		errlog.Println("Couldn't serialize configuration.")
	}


	return config
}
