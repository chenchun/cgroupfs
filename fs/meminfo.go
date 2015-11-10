package fs

import (
	"fmt"

	"bazil.org/fuse"
	_ "bazil.org/fuse/fs/fstestutil"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
)

//http://man7.org/linux/man-pages/man5/proc.5.html /proc/meminfo fs/proc/meminfo.c
//MemTotal:       %8lu kB
//MemFree:        %8lu kB
//MemAvailable:   %8lu kB
//Buffers:        %8lu kB
//Cached:         %8lu kB
//SwapCached:     %8lu kB
//Active:         %8lu kB
//Inactive:       %8lu kB
//Active(anon):   %8lu kB
//Inactive(anon): %8lu kB
//Active(file):   %8lu kB
//Inactive(file): %8lu kB
//Unevictable:    %8lu kB
//Mlocked:        %8lu kB
//SwapTotal:      %8lu kB
//SwapFree:       %8lu kB
//Dirty:          %8lu kB
//Writeback:      %8lu kB
//AnonPages:      %8lu kB
//Mapped:         %8lu kB
//Shmem:          %8lu kB
//Slab:           %8lu kB
//SReclaimable:   %8lu kB
//SUnreclaim:     %8lu kB
//KernelStack:    %8lu kB
//PageTables:     %8lu kB
//NFS_Unstable:   %8lu kB
//Bounce:         %8lu kB
//WritebackTmp:   %8lu kB
//CommitLimit:    %8lu kB
//Committed_AS:   %8lu kB
//VmallocTotal:   %8lu kB
//VmallocUsed:    %8lu kB
//VmallocChunk:   %8lu kB
const content = `MemTotal:       %d kB
MemFree:        %d kB
MemAvailable:   %d kB
Buffers:        %d kB
Cached:         %d kB
SwapCached:     %d kB
`

//Buffers:        %s kB
//Cached:         %s kB
//SwapCached:     %s kB
//Active:         %s kB
//Inactive:       %s kB
//Active(anon):   %s kB
//Inactive(anon): %s kB
//Active(file):   %s kB
//Inactive(file): %s kB
//Unevictable:    %s kB
//Mlocked:        %s kB
//SwapTotal:      %s kB
//SwapFree:       %s kB

var (
	// https://www.kernel.org/doc/Documentation/cgroups/cgroups.txt
	hardLimit   = "memory.limit_in_bytes"
	softLimit   = "memory.soft_limit_in_bytes"
	swapLimit   = "memory.memsw.limit_in_bytes"
	kernelLimit = "memory.kmem.limit_in_bytes"
	oomControl  = "memory.oom_control"
	swappniess  = "memory.swappiness"
	memusage    = "memory.usage_in_bytes"
)

type MemInfoFile struct {
	cgroupdir string
	memCgroup fs.MemoryGroup
}

func NewMemInfoFile(cgroupdir string) MemInfoFile {
	return MemInfoFile{cgroupdir, fs.MemoryGroup{}}
}

func (mi MemInfoFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0444
	data, _ := mi.ReadAll(ctx)
	a.Size = uint64(len(data))
	return nil
}

func (mi MemInfoFile) ReadAll(ctx context.Context) ([]byte, error) {
	stats := cgroups.NewStats()
	mi.memCgroup.GetStats(mi.cgroupdir, stats)
	memStats := stats.MemoryStats
	mls := mi.getLimits()
	memInfo := fmt.Sprintf(content,
		mls[hardLimit]/1024,
		(mls[hardLimit] - memStats.Usage.Usage)/1024,
		(mls[hardLimit] - memStats.Usage.Usage)/1024,
		0,
		memStats.Stats["total_cache"]/1024,
		0,
	)
	logrus.Debugf("memInfo \n%s", memInfo)
	return []byte(memInfo), nil
}

func (mi MemInfoFile) getLimits() map[string]uint64 {
	m := make(map[string]uint64)
	for _, file := range []string{
		hardLimit,
		softLimit,
		swapLimit,
		kernelLimit,
	} {
		val, err := getCgroupParamUint(mi.cgroupdir, file)
		if err != nil {
			logrus.Debugf("Error getting memory %s:%v", file, err)
		} else {
			m[file] = val
		}
	}
	return m
}
