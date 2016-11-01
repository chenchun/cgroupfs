package fs

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
	"golang.org/x/net/context"
)

// File implements both Node and Handle for the hello file.
type DiskStatsFile struct {
	cgroupdir  string
	blkioGroup fs.BlkioGroup
}

const (
	DiskStatName = "diskstats"
)

func init() {
	fileMap[DiskStatName] = &FileInfo{
		initFunc:   NewDiskStatsFile,
		inode:      INODE_DISKSTATS,
		subsysName: "blkio",
	}
}

func NewDiskStatsFile(cgroupdir string, info *FileInfo) {
	info.node = DiskStatsFile{cgroupdir, fs.BlkioGroup{}}
}

//https://www.kernel.org/doc/Documentation/cgroups/blkio-controller.txt
//https://www.kernel.org/doc/Documentation/iostats.txt

//blkio.throttle.io_service_bytes
//8:0 Read 0
//8:0 Write 774144
//8:0 Sync 0
//8:0 Async 774144
//8:0 Total 774144
//Total 774144
//blkio.throttle.io_serviced
//8:0 Read 0
//8:0 Write 189
//8:0 Sync 0
//8:0 Async 189
//8:0 Total 189
//Total 189

//docker create  --device /dev/fuse --cap-add SYS_ADMIN -v /tmp/cgroupfs/meminfo:/proc/meminfo ubuntu bash -c "while true; do echo 213 > /tmp/log; sleep 1; cat /tmp/log; done"

type BlkioOp string

const (
	Read  BlkioOp = "Read"
	Write BlkioOp = "Write"
	Sync  BlkioOp = "Sync"
	Async BlkioOp = "Async"
	Total BlkioOp = "Total"
)

func (ds DiskStatsFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = INODE_DISKSTATS
	a.Mode = 0444
	data, _ := ds.readAll()
	a.Size = uint64(len(data))
	return nil
}

func (ds DiskStatsFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fusefs.Handle, error) {
	resp.Flags |= fuse.OpenDirectIO
	return ds, nil
}

func (ds DiskStatsFile) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	data, _ := ds.readAll()
	fuseutil.HandleRead(req, resp, data)
	return nil
}

func (ds DiskStatsFile) readAll() ([]byte, error) {
	stats := cgroups.NewStats()
	ds.blkioGroup.GetStats(ds.cgroupdir, stats)

	diskStats, err := ioutil.ReadFile("/proc/diskstats")
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read /proc/diskstats: %v", err)
	}
	return getDiskStats(diskStats, stats.BlkioStats), nil
}

func getDiskStats(diskStats []byte, blkioStats cgroups.BlkioStats) []byte {
	sc := bufio.NewScanner(bytes.NewReader(diskStats))
	var (
		major, minor                  uint64
		dev                           string
		read, readMerged              uint64
		readSector, readMiliSec       uint64
		write, writeMerged            uint64
		writeSector, writeMiliSec     uint64
		ios, ioMiliSec, ioWaitMiliSec uint64
	)
	buf := bytes.NewBuffer(make([]byte, 0))
	for sc.Scan() {
		line := sc.Text()
		if n, err := fmt.Sscanf(line, "%d %d %s", &major, &minor, &dev); n == 3 && err == nil {
			logrus.Debugf("Get blkio stats %d %d %s", major, minor, dev)
			read = getBlkioStats(blkioStats.IoServicedRecursive, major, minor, Read)
			readMerged = getBlkioStats(blkioStats.IoMergedRecursive, major, minor, Read)
			readSector = getBlkioStats(blkioStats.IoServiceBytesRecursive, major, minor, Read) / 512
			readMiliSec = (getBlkioStats(blkioStats.IoServiceTimeRecursive, major, minor, Read) + getBlkioStats(blkioStats.IoWaitTimeRecursive, major, minor, Read)) / 1000000
			write = getBlkioStats(blkioStats.IoServicedRecursive, major, minor, Write)
			writeMerged = getBlkioStats(blkioStats.IoMergedRecursive, major, minor, Write)
			writeSector = getBlkioStats(blkioStats.IoServiceBytesRecursive, major, minor, Write) / 512
			writeMiliSec = (getBlkioStats(blkioStats.IoServiceTimeRecursive, major, minor, Write) + getBlkioStats(blkioStats.IoWaitTimeRecursive, major, minor, Write)) / 1000000
			ios = getBlkioStats(blkioStats.IoQueuedRecursive, major, minor, Total)
			ioMiliSec = getBlkioStats(blkioStats.IoTimeRecursive, major, minor, Total) / 1000000
			ioWaitMiliSec = getBlkioStats(blkioStats.IoWaitTimeRecursive, major, minor, Total) / 1000000
			if read != 0 || readMerged != 0 || readSector != 0 || readMiliSec != 0 || write != 0 || writeMerged != 0 || writeSector != 0 || writeMiliSec != 0 || ios != 0 || ioMiliSec != 0 || ioWaitMiliSec != 0 {
				fmt.Fprintf(buf, "%d       %d %s %d %d %d %d %d %d %d %d %d %d %d\n",
					major, minor, dev, read, readMerged, readSector, readMiliSec,
					write, writeMerged, writeSector, writeMiliSec, ios, ioMiliSec, ioWaitMiliSec)
			}
		}
	}
	return buf.Bytes()
}

func getBlkioStats(entries []cgroups.BlkioStatEntry, major, minor uint64, op BlkioOp) uint64 {
	for _, entry := range entries {
		if entry.Major == major && entry.Minor == minor && entry.Op == string(op) {
			return entry.Value
		}
	}
	return 0
}
