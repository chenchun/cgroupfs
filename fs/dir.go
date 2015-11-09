package fs

import (
	"os"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"golang.org/x/net/context"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"path/filepath"
)

// Dir implements both Node and Handle for the root directory.
type Dir struct{
	cgroupdir string
}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	if name == "hello" {
		return File{}, nil
	} else if name == "meminfo" {
		memMountPoint, err := cgroups.FindCgroupMountpoint("memory")
		if err != nil {
			return nil, fuse.ENODATA
		}
		return NewMemInfoFile(filepath.Join(memMountPoint, d.cgroupdir)), nil
	}
	return nil, fuse.ENOENT
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{Inode: 2, Name: "hello", Type: fuse.DT_File},
		{Inode: 3, Name: "meminfo", Type: fuse.DT_File},
	}, nil
}