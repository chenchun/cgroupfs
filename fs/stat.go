package fs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
)

type StatFile struct {
	cgroupdir string
}

var (
	statModifier *regexp.Regexp = nil
	statSep      *regexp.Regexp = nil
)

const (
	MAXFIELDS   = 11
	CpuStatName = "stat"
)

func NewStatFile(cgroupdir string, info *FileInfo) {
	info.node = StatFile{cgroupdir}
}

func (sf StatFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = INODE_STAT
	a.Mode = 0444
	data, _ := sf.readAll()
	a.Size = uint64(len(data))

	return nil
}

func (sf StatFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	resp.Flags |= fuse.OpenDirectIO
	return sf, nil
}

func (sf StatFile) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data, _ := sf.readAll()
	fuseutil.HandleRead(req, resp, data)
	return nil
}

func (sf StatFile) readAll() ([]byte, error) {
	var buffer bytes.Buffer

	if statModifier != nil {
		sf.getStatInfo(&buffer, getCpuSets(sf.cgroupdir))
	}

	return buffer.Bytes(), nil
}

func (sf StatFile) getStatInfo(buffer *bytes.Buffer, cpuIDs map[uint64]uint64) {
	buffer.Reset()

	if cpuIDs == nil {
		return
	}

	rawContent, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}

	var (
		count     int = 0
		cpuStat       = make([]uint64, MAXFIELDS-1)
		num       uint64
		tmpBuffer bytes.Buffer
	)

	for _, line := range strings.Split(string(rawContent), "\n") {
		groups := statModifier.FindSubmatch([]byte(line))
		if len(groups) == 2 {
			// we do not check the error after calling parseUnit, because
			// kernel guarantees for us
			if len(groups[1]) == 0 {
				continue
			}

			cpuID, _ := parseUint(string(groups[1]), 10, 32)
			if _, ok := cpuIDs[cpuID]; ok {
				for i, item := range statSep.Split(line, MAXFIELDS)[1:] {
					num, _ = parseUint(item, 10, 64)
					cpuStat[i] += num
				}
				tmpBuffer.WriteString(statModifier.ReplaceAllString(line, fmt.Sprintf("cpu%d", count)))
				tmpBuffer.WriteString("\n")
				count++
			}
		} else {
			tmpBuffer.WriteString(line)
			tmpBuffer.WriteString("\n")
		}
	}

	buffer.WriteString("cpu ")
	for _, item := range cpuStat {
		buffer.WriteString(" ")
		buffer.WriteString(strconv.FormatUint(item, 10))
	}
	buffer.WriteString("\n")
	buffer.Write(tmpBuffer.Bytes())
}

func init() {
	if runtime.GOOS == "linux" {
		fileMap[CpuStatName] = &FileInfo{
			initFunc:   NewStatFile,
			inode:      INODE_STAT,
			subsysName: "cpuset",
		}

		statModifier, _ = regexp.Compile("cpu(\\d*)")
		statSep, _ = regexp.Compile("\\s+")
	}
}
