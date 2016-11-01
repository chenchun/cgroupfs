package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
)

func TestDiskStats(t *testing.T) {
	stats := cgroups.NewStats()
	pkgPath := packagePath()
	if pkgPath == "" {
		t.Skip("failed to find pkg path")
	}
	hackDir := filepath.Join(pkgPath, "src/github.com/chenchun/cgroupfs/hack/")
	blkioCgroup := &fs.BlkioGroup{}
	blkioCgroup.GetStats(filepath.Join(hackDir, "blkio"), stats)

	diskStats, err := ioutil.ReadFile(filepath.Join(hackDir, "/proc/diskstats"))
	if err != nil {
		t.Fatalf("failed to read /proc/diskstats: %v", err)
	}
	ret := string(getDiskStats(diskStats, stats.BlkioStats))
	if ret != "8       0 sda 137 1 1608 730 377 3 3392 263 0 0 688\n" {
		t.Fatalf("%q", ret)
	}
}

func packagePath() string {
	gopaths := os.Getenv("GOPATH")
	for _, path := range strings.Split(gopaths, ":") {
		if strings.Contains(path, "Godeps") {
			continue
		}
		return path
	}
	return ""
}
