package cgroupfs

import (
	"testing"
	"os"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"time"
	"syscall"
	"strings"

	"github.com/opencontainers/runc/libcontainer/cgroups"
)

func TestMemory(t *testing.T) {
	helper, err := newCgroupfsHelper("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := helper.apply("memory", "memory.limit_in_bytes", "102400"); err != nil {
		t.Fatal(err)
	}
	if !helper.startCgroupfs(3*time.Second) {
		t.Fatal("time out waiting for cgroupfs to start")
	}
	if data, err := helper.readAll("meminfo"); err != nil {
		t.Fatal(err)
	} else {
		expect := `MemTotal:       100 kB
MemFree:        100 kB
MemAvailable:   100 kB
Buffers:        0 kB
Cached:         0 kB
SwapCached:     0 kB
`
		if string(data) != expect {
			t.Fatalf("content mismatch %s", string(data))
		}
	}
	if err := helper.stopCgroupfs(); err != nil {
		t.Fatal(err)
	}
}

type cgroupfsHelper struct {
	mountpoint string
	cgroupDir string
}

func newCgroupfsHelper(mountpoint, cgroupDir string) (*cgroupfsHelper, error) {
	var err error
	if mountpoint == "" {
		if mountpoint, err = ioutil.TempDir("", ""); err != nil {
			return nil, err
		}
	}
	memCgroupDir, err := cgroups.FindCgroupMountpoint("memory")
	if err != nil {
		return nil, err
	}
	if cgroupDir == "" {
		if cgroupDir, err = ioutil.TempDir(memCgroupDir, ""); err != nil {
			return nil, err
		}
		cgroupDir = strings.TrimPrefix(cgroupDir, memCgroupDir)
	}
	return &cgroupfsHelper{mountpoint, cgroupDir}, nil
}

func (h *cgroupfsHelper) apply(subsystem, file, data string) error {
	subsystemDir, err := cgroups.FindCgroupMountpoint(subsystem)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(subsystemDir, h.cgroupDir), 0700); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(subsystemDir, h.cgroupDir), file, data); err != nil {
		return err
	}
	return nil
}

func (h *cgroupfsHelper) waitForStart(timeout time.Duration) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	select {
	case <-ticker.C:
		if exist(filepath.Join(h.mountpoint, "meminfo")) {
			return true
		}
	case <-time.After(timeout):
		break
	}
	return false
}

func (h *cgroupfsHelper) startCgroupfs(timeout time.Duration) bool {
	go Serve(h.mountpoint, h.cgroupDir)
	return h.waitForStart(timeout)
}

func (h *cgroupfsHelper) stopCgroupfs() error {
	return syscall.Unmount(h.mountpoint, 0)
}

func (h *cgroupfsHelper) readAll(file string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(h.mountpoint, file))
}

func writeFile(dir, file, data string) error {
	// Normally dir should not be empty, one case is that cgroup subsystem
	// is not mounted, we will get empty dir, and we want it fail here.
	if dir == "" {
		return fmt.Errorf("no such directory for %s.", file)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, file), []byte(data), 0700); err != nil {
		return fmt.Errorf("failed to write %v to %v: %v", data, file, err)
	}
	return nil
}

func exist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
