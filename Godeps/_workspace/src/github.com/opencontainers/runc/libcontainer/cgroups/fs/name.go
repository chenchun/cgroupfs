// +build linux

package fs

import (
	"github.com/chenchun/cgroupfs/Godeps/_workspace/src/github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/chenchun/cgroupfs/Godeps/_workspace/src/github.com/opencontainers/runc/libcontainer/configs"
)

type NameGroup struct {
	GroupName string
}

func (s *NameGroup) Name() string {
	return s.GroupName
}

func (s *NameGroup) Apply(d *data) error {
	return nil
}

func (s *NameGroup) Set(path string, cgroup *configs.Cgroup) error {
	return nil
}

func (s *NameGroup) Remove(d *data) error {
	return nil
}

func (s *NameGroup) GetStats(path string, stats *cgroups.Stats) error {
	return nil
}
