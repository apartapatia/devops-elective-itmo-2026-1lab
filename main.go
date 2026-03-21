package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Version  string    `json:"ociVersion"` 
	Process  Process   `json:"process"`    
	Root     Root      `json:"root"`       
	Hostname string    `json:"hostname"`   
	Mounts   []Mount   `json:"mounts"`     
	Linux    Linux     `json:"linux"`      
}

type Process struct {
	Terminal bool `json:"terminal"`
	Cwd 	 string `json:"cwd"`
	Args 	 []string `json:"args"`
	Env 	 []string `json:"env"`
	
	User struct {
			UID uint `json:"uid"`
			GID uint `json:"gid"`
		} `json:"user"`
		
	Capabilities struct {
			Bounding []string `json:"bounding"`
			Effective []string `json:"effective"`
			Permitted []string `json:"permitted"`
		} `json:"capabilities"`
	
	Rlimits []struct {
			Type string `json:"type"`
			Hard uint64 `json:"hard"`
			Soft uint64 `json:"soft"`
	}	`json:"rlimits"`
		
}

type Root struct {
	Path     string `json:"path"`
	Readonly bool   `json:"readonly"`
}

type Mount struct {
	Destination string   `json:"destination"`
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Options     []string `json:"options"` 
}

type Linux struct {
	Namespaces []struct {
		Type string `json:"type"`
	} `json:"namespaces"`
}

func main() {
	configJson, err := os.Open("config.json")
	if err != nil {
		fmt.Printf("Ошибка чтения config.json: %v\n", err)
		os.Exit(1)
	}
	defer configJson.Close()
	
	var config Config
	
	err = json.NewDecoder(configJson).Decode(&config)
	if err != nil {
		fmt.Printf("Ошибка парсинга config.json: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Успешное открытие конфига 🎉")
	fmt.Printf("HOSTNAME: %s\n", config.Hostname)
	fmt.Printf("EXECUTABLE PROGRAMM: %v\n", config.Process.Args)
}