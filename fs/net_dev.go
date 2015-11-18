package fs

import (
	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"bytes"
	"golang.org/x/net/context"
	"io/ioutil"
	"strings"
)

const NET_DEV_FILE = "/proc/net/dev"

var (
	buffer bytes.Buffer
)

type NetDevFile struct {
	vethName string
}

func init() {
	fileMap["net_dev"] = FileInfo{
		initFunc:   NewNetDevFile,
		inode:      INODE_NET_DEV,
		subsysName: "",
	}
}

func NewNetDevFile(vethName string) fusefs.Node {
	return NetDevFile{vethName}
}

func (nd NetDevFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = INODE_NET_DEV
	a.Mode = 0444
	data, _ := nd.ReadAll(ctx)
	a.Size = uint64(len(data))
	return nil
}

func (nd NetDevFile) ReadAll(ctx context.Context) ([]byte, error) {
	netDev, err := getNetDevFromNetFile(nd.vethName)
	return []byte(netDev), err
}

func getNetDevFromNetFile(vethName string) (string, error) {
	rawContent, err := ioutil.ReadFile(NET_DEV_FILE)
	if err != nil {
		return "", err
	}

	buffer.Reset()
	for index, line := range strings.Split(string(rawContent), "\n") {
		// skip empty line
		if len(line) == 0 {
			continue
		}
		// read head of title
		if index == 0 || index == 1 {
			buffer.WriteString(line)
			buffer.WriteString("\n")
		}
		if len(vethName) != 0 && strings.HasPrefix(line, vethName) {
			buffer.WriteString(replaceVethNameWithEth0(line, vethName))
			buffer.WriteString("\n")
			break
		}
	}

	return buffer.String(), nil
}

func replaceVethNameWithEth0(content, vethName string) string {
	return strings.Replace(content, vethName, "  eth0", 1)
}
