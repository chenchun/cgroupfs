// CgroupFS implements a simple "hello world" file system.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/chenchun/cgroupfs/Godeps/_workspace/src/bazil.org/fuse"
	fusefs "github.com/chenchun/cgroupfs/Godeps/_workspace/src/bazil.org/fuse/fs"
	_ "github.com/chenchun/cgroupfs/Godeps/_workspace/src/bazil.org/fuse/fs/fstestutil"

	"github.com/chenchun/cgroupfs/fs"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT CGROUP_DIR\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if flag.NArg() != 2 {
		Usage()
		os.Exit(2)
	}
	mountpoint := flag.Arg(0)
	cgroupdir := flag.Arg(1)

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("cgroupfs"),
		fuse.Subtype("cgroupfs"),
		fuse.LocalVolume(),
		fuse.VolumeName("cgroup volume"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	err = fusefs.Serve(c, fs.FS{cgroupdir})
	if err != nil {
		log.Fatal(err)
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}
