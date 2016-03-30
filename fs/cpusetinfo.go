package fs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type CpuInfoFile struct {
	cgroupdir string
}

var (
	cpuinfoModifier *regexp.Regexp = nil
)

func NewCpuInfoFile(cgroupdir string) fusefs.Node {
	return CpuInfoFile{cgroupdir}
}

func (ci CpuInfoFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = INODE_CPUINFO
	a.Mode = 0444
	data, _ := ci.ReadAll(ctx)
	a.Size = uint64(len(data))

	return nil
}

func (ci CpuInfoFile) ReadAll(ctx context.Context) ([]byte, error) {
	var buffer bytes.Buffer

	if cpuinfoModifier != nil {
		ci.getCpuInfo(&buffer, getCpuSets(ci.cgroupdir))
	}

	return buffer.Bytes(), nil
}

func (ci CpuInfoFile) getCpuInfo(buffer *bytes.Buffer, cpuIDs map[uint64]uint64) {
	if cpuIDs == nil {
		return
	}

	rawContent, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		return
	}

	count := 0
	for _, line := range strings.Split(string(rawContent), "\n\n") {
		groups := cpuinfoModifier.FindSubmatch([]byte(line))
		if len(groups) == 2 {
			// we do not check the error after calling parseUnit, because
			// kernel guarantees for us
			cpuID, _ := parseUint(string(groups[1]), 10, 32)
			if _, ok := cpuIDs[cpuID]; ok {
				buffer.WriteString(cpuinfoModifier.ReplaceAllString(line, fmt.Sprintf("%-16s: %d", "processor", count)))
				buffer.WriteString("\n\n")
				count++
			}
		}
	}
}

func init() {
	if runtime.GOOS == "linux" {
		fileMap["cpuinfo"] = FileInfo{
			initFunc:   NewCpuInfoFile,
			inode:      INODE_CPUINFO,
			subsysName: "cpuset",
		}

		cpuinfoModifier, _ = regexp.Compile("processor\\s+?:\\s+?(\\d+)")
	}
}
