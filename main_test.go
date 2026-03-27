package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ContainerTestSuite struct {
	suite.Suite
	containerCmd *exec.Cmd
	containerPid int
	containerId  string
	stdinPipe    io.WriteCloser
}

func (s *ContainerTestSuite) SetupSuite() {
	s.containerCmd = exec.Command("./apartapatia-runtime", "run")
	s.containerCmd.Stdout = io.Discard
	s.containerCmd.Stderr = io.Discard

	var err error
	s.stdinPipe, err = s.containerCmd.StdinPipe()
	require.NoError(s.T(), err, "не удалось создать stdin pipe")

	err = s.containerCmd.Start()
	require.NoError(s.T(), err, "не удалось запустить runtime")

	s.T().Logf("[setup] parent PID=%d", s.containerCmd.Process.Pid)

	found := assert.Eventually(s.T(), func() bool {
		pid, id, err := findContainerChildProcess(s.containerCmd.Process.Pid)
		if err != nil || pid == 0 {
			return false
		}
		s.containerPid = pid
		s.containerId = id
		return true
	}, 5*time.Second, 100*time.Millisecond)

	require.True(s.T(), found, "child не найден после старта")
	s.T().Logf("[setup] child PID=%d, containerID=%s", s.containerPid, s.containerId)
}

func (s *ContainerTestSuite) TearDownSuite() {
	s.T().Log("[teardown] очистка ресурсов...")

	if s.stdinPipe != nil {
		s.stdinPipe.Close()
	}

	if s.containerCmd != nil && s.containerCmd.Process != nil {
		_ = s.containerCmd.Process.Kill()
		_ = s.containerCmd.Wait()
		s.T().Logf("[teardown] процесс PID=%d завершён", s.containerCmd.Process.Pid)
	}

	if s.containerId != "" {
		basePath := filepath.Join("/var/lib/apartapatia-runtime", s.containerId)
		mergedDir := filepath.Join(basePath, "merged")
		procInContainer := filepath.Join(mergedDir, "proc")

		if err := syscall.Unmount(procInContainer, 0); err != nil {
			s.T().Logf("[teardown] /proc: %v", err)
		} else {
			s.T().Log("[teardown] /proc размонтирован")
		}

		if err := syscall.Unmount(mergedDir, syscall.MNT_DETACH); err != nil {
			s.T().Logf("[teardown] overlay: %v", err)
		} else {
			s.T().Log("[teardown] overlay размонтирован")
		}

		if err := os.RemoveAll(basePath); err != nil {
			s.T().Logf("[teardown] НЕ УДАЛОСЬ УДАЛИТЬ %s: %v", basePath, err)
		} else {
			s.T().Logf("[teardown] директория %s удалена", basePath)
		}
	}

	s.T().Log("[teardown] очистка завершена")
}

func (s *ContainerTestSuite) TestNamespaces() {
	tests := []struct {
		name   string
		nsType string
	}{
		{"UTS namespace изолирован от хоста", "uts"},
		{"PID namespace изолирован от хоста", "pid"},
		{"Mount namespace изолирован от хоста", "mnt"},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			hostNs, err := readNamespaceSymlink(1, tt.nsType)
			require.NoError(s.T(), err, "не удалось прочитать %s namespace хоста", tt.nsType)

			containerNs, err := readNamespaceSymlink(s.containerPid, tt.nsType)
			require.NoError(s.T(), err, "не удалось прочитать %s namespace контейнера", tt.nsType)

			s.T().Logf("[%s] host=%s  container=%s", tt.nsType, hostNs, containerNs)

			assert.NotEqual(s.T(), hostNs, containerNs,
				"%s должен быть изолирован: host=%s container=%s", tt.name, hostNs, containerNs)
		})
	}
}

func (s *ContainerTestSuite) TestContainerDirectoryStructure() {
	require.NotEmpty(s.T(), s.containerId, "containerId не определён. ошибка нахождения дочернего процесса")

	containerPath := filepath.Join("/var/lib/apartapatia-runtime", s.containerId)
	s.T().Logf("[dirs] проверяем %s", containerPath)

	s.Run("upper layer существует", func() {
		assert.DirExists(s.T(), filepath.Join(containerPath, "upper"), "upperdir должен быть создан")
	})

	s.Run("work layer существует", func() {
		assert.DirExists(s.T(), filepath.Join(containerPath, "work"), "workdir должен быть создан")
	})

	s.Run("merged layer существует", func() {
		assert.DirExists(s.T(), filepath.Join(containerPath, "merged"), "mergeddir должен быть создан")
	})
}

func (s *ContainerTestSuite) TestProcMount() {
	procPath := filepath.Join("/proc", strconv.Itoa(s.containerPid), "root", "proc")
	s.T().Logf("[proc] проверяем %s", procPath)

	assert.DirExists(s.T(), procPath, "/proc должен существовать внутри контейнера")
	assert.FileExists(s.T(), filepath.Join(procPath, "1", "cmdline"), "/proc должен быть смонтирован: ожидается /proc/1/cmdline")
}

func (s *ContainerTestSuite) TestPID1() {
	cmdlinePath := filepath.Join("/proc", strconv.Itoa(s.containerPid), "root", "proc", "1", "cmdline")

	cmdline, err := os.ReadFile(cmdlinePath)
	if !assert.NoError(s.T(), err, "не удалось прочитать /proc/1/cmdline") {
		return
	}

	processName := filepath.Base(strings.SplitN(string(cmdline), "\x00", 2)[0])
	s.T().Logf("[pid1] процесс: %q", processName)

	assert.Equal(s.T(), "sh", processName, "PID=1 должен быть 'sh'")
}

func (s *ContainerTestSuite) TestHostname() {
	cmd := exec.Command("nsenter", "-t", strconv.Itoa(s.containerPid), "-u", "hostname")
	out, err := cmd.CombinedOutput()
	s.T().Logf("[hostname] nsenter: %q, err=%v", strings.TrimSpace(string(out)), err)

	if !assert.NoError(s.T(), err, "nsenter должен выполниться успешно") {
		return
	}

	assert.Equal(s.T(), "apartapatia-runtime", strings.TrimSpace(string(out)),
		"hostname должен совпадать со значением из config.json")
}

func TestContainerSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}

func readNamespaceSymlink(pid int, nsType string) (string, error) {
	return os.Readlink(filepath.Join("/proc", strconv.Itoa(pid), "ns", nsType))
}

func findContainerChildProcess(parentPid int) (int, string, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, "", fmt.Errorf("не удалось прочитать /proc: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 1 {
			continue
		}

		statBytes, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "stat"))
		if err != nil {
			continue
		}

		idx := strings.LastIndex(string(statBytes), ") ")
		if idx == -1 {
			continue
		}
		fields := strings.Fields(string(statBytes[idx+2:]))
		if len(fields) < 2 {
			continue
		}
		ppid, _ := strconv.Atoi(fields[1])
		if ppid != parentPid {
			continue
		}

		cwd, err := os.Readlink(filepath.Join("/proc", entry.Name(), "cwd"))
		if err != nil {
			return pid, "", nil
		}

		return pid, extractContainerID(cwd), nil
	}

	return 0, "", nil
}

func extractContainerID(cwd string) string {
	const prefix = "/var/lib/apartapatia-runtime/"
	if !strings.HasPrefix(cwd, prefix) {
		return ""
	}
	parts := strings.SplitN(strings.TrimPrefix(cwd, prefix), "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
