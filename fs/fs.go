package fs

import (
	"sync"

	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
)

var (
	files     []string
	filesOnce sync.Once
)

// FS implements the hello world file system.
type FS struct {
	CgroupDir string
}

func (fs FS) Root() (fs.Node, error) {
	return Dir{fs.CgroupDir}, nil
}

func ProcFiles() []string {
	filesOnce.Do(func() {
		for k := range fileMap {
			files = append(files, k)
		}
	})
	return files
}
