package fs

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
	"golang.org/x/net/context"
)

// File implements both Node and Handle for the vmstat file.
type VMStatFile struct {
	cgroupdir  string
	blkioGroup fs.BlkioGroup
}

const (
	VMStatName = "vmstat"
)

func init() {
	fileMap[VMStatName] = &FileInfo{
		initFunc:   NewVMStatFile,
		inode:      INODE_VMSTAT,
		subsysName: "blkio",
	}
}

func NewVMStatFile(cgroupdir string, info *FileInfo) {
	info.node = VMStatFile{cgroupdir, fs.BlkioGroup{}}
}

func (v VMStatFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = INODE_VMSTAT
	a.Mode = 0444
	data, _ := v.readAll()
	a.Size = uint64(len(data))
	return nil
}

func (v VMStatFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fusefs.Handle, error) {
	resp.Flags |= fuse.OpenDirectIO
	return v, nil
}

func (v VMStatFile) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data, _ := v.readAll()
	fuseutil.HandleRead(req, resp, data)
	return nil
}

func (v VMStatFile) readAll() ([]byte, error) {
	stats := cgroups.NewStats()
	v.blkioGroup.GetStats(v.cgroupdir, stats)
	vmstat, err := ioutil.ReadFile("/proc/vmstat")
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read /proc/vmstat: %v", err)
	}
	return getVMStat(vmstat, stats.BlkioStats)
}

func getVMStat(vmstat []byte, blkioStats cgroups.BlkioStats) ([]byte, error) {
	var (
		read, write uint64
		err         error
	)
	for _, entry := range blkioStats.IoServiceBytesRecursive {
		if entry.Op == string(Read) {
			read += entry.Value
		} else if entry.Op == string(Write) {
			write += entry.Value
		}
	}
	buf := bytes.NewBuffer(make([]byte, 0))
	sc := bufio.NewScanner(bytes.NewReader(vmstat))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "pgpgin") {
			_, err = fmt.Fprintf(buf, "pgpgin %d\n", read/1024)
		} else if strings.HasPrefix(line, "pgpgout") {
			_, err = fmt.Fprintf(buf, "pgpgout %d\n", write/1024)
		} else {
			_, err = fmt.Fprintf(buf, "%s\n", line)
		}
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
