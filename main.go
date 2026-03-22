package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func generateID() string {
	b := make([]byte, 6)
	rand.Read(b)
	
	return hex.EncodeToString(b)
}

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

func stdConnect(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

func run(config Config, containerId string) {
	exeLinuxPath, err := os.Executable()
	if err  != nil {
		fmt.Printf("Ошибка получаения exe path: %v\n", err)
		os.Exit(1)
	}
	
	args := append([]string{"child", containerId}, config.Process.Args...)
	
	cmd := exec.Command(exeLinuxPath, args...)
	
	stdConnect(cmd)
	
	cmd.SysProcAttr = &syscall.SysProcAttr {
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Ошибка запуска parent процесса: %v\n", err)
		os.Exit(1)
	}
}

func child(config Config, containerId string) { // containerId зачем тут он нужен?
	cmd := exec.Command(config.Process.Args[0], config.Process.Args[1:]...)
	
	stdConnect(cmd)
	
	syscall.Sethostname([]byte(config.Hostname))
	if err := cmd.Run(); err != nil {
		fmt.Printf("Ошибка запуска child процесса: %v\n", err)
		os.Exit(1)
	}
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
	
	switch os.Args[1] {
		case "run":
		    containerId := generateID()
			run(config, containerId)
		case "child":
		    containerId := os.Args[2]
			child(config, containerId)
		default:
			panic("я упал")
		}
}