package fs

import (
	"os"
	"path/filepath"
	"sync"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"

	"github.com/opencontainers/runc/libcontainer/cgroups"

	"golang.org/x/net/context"
)

const (
	_ = iota
	INODE_DIR
	INODE_HELLO
	INODE_MEMINFO
	INODE_DISKSTATS
	INODE_CPUINFO
	INODE_STAT
	INODE_NET_DEV
)

var (
	fileMap = make(map[string]FileInfo)

	direntsOnce sync.Once
	dirents     []fuse.Dirent
)

// Dir implements both Node and Handle for the root directory.
type Dir struct {
	cgroupdir string
	vethName  string
}

type FileInfo struct {
	initFunc   func(cgroupdir string) fusefs.Node
	inode      uint64
	subsysName string
}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = INODE_DIR
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	if name == "hello" {
		return File{}, nil
	} else if name == "net_dev" {
		if fileInfo, ok := fileMap[name]; ok {
			return fileInfo.initFunc(d.vethName), nil
		}
	} else if fileInfo, ok := fileMap[name]; ok {
		mountPoint, err := cgroups.FindCgroupMountpoint(fileInfo.subsysName)
		if err != nil {
			return nil, fuse.ENODATA
		}
		return fileInfo.initFunc(filepath.Join(mountPoint, d.cgroupdir)), nil
	}
	return nil, fuse.ENOENT
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	direntsOnce.Do(func() {
		dirents = append(dirents, fuse.Dirent{Inode: INODE_HELLO, Name: "hello", Type: fuse.DT_File})
		for k, v := range fileMap {
			dirents = append(dirents, fuse.Dirent{Inode: v.inode, Name: k, Type: fuse.DT_File})
		}
	})
	return dirents, nil
}
