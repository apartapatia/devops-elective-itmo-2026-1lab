package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func generateID() string {
	b := make([]byte, 6)
	rand.Read(b)

	return hex.EncodeToString(b)
}

type Config struct {
	Version  string  `json:"ociVersion"`
	Process  Process `json:"process"`
	Root     Root    `json:"root"`
	Hostname string  `json:"hostname"`
	Mounts   []Mount `json:"mounts"`
	Linux    Linux   `json:"linux"`
}

type Process struct {
	Terminal bool     `json:"terminal"`
	Cwd      string   `json:"cwd"`
	Args     []string `json:"args"`
	Env      []string `json:"env"`

	User struct {
		UID uint `json:"uid"`
		GID uint `json:"gid"`
	} `json:"user"`

	Capabilities struct {
		Bounding  []string `json:"bounding"`
		Effective []string `json:"effective"`
		Permitted []string `json:"permitted"`
	} `json:"capabilities"`

	Rlimits []struct {
		Type string `json:"type"`
		Hard uint64 `json:"hard"`
		Soft uint64 `json:"soft"`
	} `json:"rlimits"`
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
	if err != nil {
		fmt.Printf("exe path не поулчен: %v\n", err)
		os.Exit(1)
	}

	args := append([]string{"child", containerId}, config.Process.Args...)

	cmd := exec.Command(exeLinuxPath, args...)

	stdConnect(cmd)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("parent процесс не запустился: %v\n", err)
		os.Exit(1)
	}
}

func child(config Config, containerId string) {
	runtimeName := "apartapatia-runtime"
	basePath := filepath.Join("/var/lib", runtimeName, containerId)

	upperDir := filepath.Join(basePath, "upper")
	workDir := filepath.Join(basePath, "work")
	mergedDir := filepath.Join(basePath, "merged")

	lowerDirAbs, err := filepath.Abs(config.Root.Path)
	if err != nil {
		fmt.Printf("rootfs не получен: %v\n", err)
		os.Exit(1)
	}

	os.MkdirAll(upperDir, 0755)
	os.MkdirAll(workDir, 0755)
	os.MkdirAll(mergedDir, 0755)

	overlayOptions := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDirAbs, upperDir, workDir)
	err = syscall.Mount("overlay", mergedDir, "overlay", 0, overlayOptions)
	if err != nil {
		fmt.Printf("overlayfs не смонтировался: %v\n", err)
		os.Exit(1)
	}

	err = syscall.Chroot(mergedDir)
	if err != nil {
		fmt.Printf("chroot отработал с ошибкой: %v\n", err)
		os.Exit(1)
	}

	err = syscall.Chdir("/")
	if err != nil {
		fmt.Printf("chdir отработал с ошибкой: %v\n", err)
		os.Exit(1)
	}

	err = syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		fmt.Printf("/proc не смонтировался: %v\n", err)
	}

	cmd := exec.Command(config.Process.Args[0], config.Process.Args[1:]...)
	stdConnect(cmd)

	syscall.Sethostname([]byte(config.Hostname))
	if err := syscall.Exec(cmd.Path, cmd.Args, os.Environ()); err != nil {
		fmt.Printf("ошибка запуска child: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "запуск контейнера через sudo")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "ошибка кол-ва аргументов: apartapatia-runtime <run|child> [container-id]")
		os.Exit(1)
	}

	configJson, err := os.Open("config.json")
	if err != nil {
		fmt.Printf("config.json не считался: %v\n", err)
		os.Exit(1)
	}
	defer configJson.Close()

	var config Config

	err = json.NewDecoder(configJson).Decode(&config)
	if err != nil {
		fmt.Printf("config.json не распарсился: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("успешное открытие конфига")
	fmt.Printf("имя хоста: %s\n", config.Hostname)
	fmt.Printf("программа для запуска: %v\n", config.Process.Args)

	switch os.Args[1] {
	case "run":
		containerId := generateID()
		fmt.Printf("Container ID: %s\n", containerId)
		run(config, containerId)
	case "child":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "не верный запуск child: apartapatia-runtime child <container-id>")
			os.Exit(1)
		}
		containerId := os.Args[2]
		child(config, containerId)
	default:
		fmt.Fprintf(os.Stderr, "неизвестная команда: %s\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "используй apartapatia-runtime <run|child>")
		os.Exit(1)
	}
}
