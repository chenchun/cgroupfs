// CgroupFS implements a simple "hello world" file system.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"

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
	if err := Serve(flag.Arg(0), flag.Arg(1)); err != nil {
		log.Fatal(err)
	}
}

func Serve(mountPoint, cgroupDir string) error {
	c, err := fuse.Mount(
		mountPoint,
		fuse.FSName("cgroupfs"),
		fuse.Subtype("cgroupfs"),
		fuse.LocalVolume(),
		fuse.VolumeName("cgroup volume"),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	err = fusefs.Serve(c, fs.FS{cgroupDir})
	if err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}
}
