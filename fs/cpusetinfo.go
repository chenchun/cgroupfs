package fs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
)

type CpuInfoFile struct {
	cgroupdir string
}

var (
	cpuinfoModifier *regexp.Regexp = nil
)

const (
	CpuInfoName = "cpuinfo"
)

func NewCpuInfoFile(cgroupdir string, info *FileInfo) {
	info.node = CpuInfoFile{cgroupdir}
}

func (ci CpuInfoFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = INODE_CPUINFO
	a.Mode = 0444
	data, _ := ci.readAll()
	a.Size = uint64(len(data))

	return nil
}

func (ci CpuInfoFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	resp.Flags |= fuse.OpenDirectIO
	return ci, nil
}

func (ci CpuInfoFile) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data, _ := ci.readAll()
	fuseutil.HandleRead(req, resp, data)
	return nil
}

func (ci CpuInfoFile) readAll() ([]byte, error) {
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
		fileMap[CpuInfoName] = &FileInfo{
			initFunc:   NewCpuInfoFile,
			inode:      INODE_CPUINFO,
			subsysName: "cpuset",
		}

		cpuinfoModifier, _ = regexp.Compile("processor\\s+?:\\s+?(\\d+)")
	}
}
