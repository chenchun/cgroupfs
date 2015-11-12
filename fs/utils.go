package fs

import (
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// From github.com/opencontainers/runc/libcontainer/cgroups/fs/utils.go
func getCgroupParamUint(cgroupPath, cgroupFile string) (uint64, error) {
	contents, err := ioutil.ReadFile(filepath.Join(cgroupPath, cgroupFile))
	if err != nil {
		return 0, err
	}

	return parseUint(strings.TrimSpace(string(contents)), 10, 64)
}

// From github.com/opencontainers/runc/libcontainer/cgroups/fs/utils.go
// Saturates negative values at zero and returns a uint64.
// Due to kernel bugs, some of the memory cgroup stats can be negative.
func parseUint(s string, base, bitSize int) (uint64, error) {
	value, err := strconv.ParseUint(s, base, bitSize)
	if err != nil {
		intValue, intErr := strconv.ParseInt(s, base, bitSize)
		// 1. Handle negative values greater than MinInt64 (and)
		// 2. Handle negative values lesser than MinInt64
		if intErr == nil && intValue < 0 {
			return 0, nil
		} else if intErr != nil && intErr.(*strconv.NumError).Err == strconv.ErrRange && intValue < 0 {
			return 0, nil
		}

		return value, err
	}

	return value, nil
}

func binarySearchInt(a []int, x int) bool {
	index := sort.SearchInts(a, x)
	return !(index == len(a)) && !(x < a[index])
}

func getCpuSets(cgrouddir string) []int {
	var (
		err               error
		rawContent        []byte
		content           string
		cpuIDs            []int
		cpuID, begin, end uint64
	)

	rawContent, err = ioutil.ReadFile(filepath.Join(cgrouddir, "cpuset.cpus"))
	if err != nil {
		return nil
	}

	content = strings.TrimSpace(string(rawContent))
	for _, split := range strings.Split(content, ",") {
		idRange := strings.Split(split, "-")
		// we do not check the error after calling parseUnit, because
		// cgroup has done it for us
		if len(idRange) == 1 {
			cpuID, _ = parseUint(idRange[0], 10, 32)
			cpuIDs = append(cpuIDs, int(cpuID))
		} else if len(idRange) == 2 {
			begin, _ = parseUint(idRange[0], 10, 32)
			end, _ = parseUint(idRange[1], 10, 32)
			for i := begin; i <= end; i++ {
				cpuIDs = append(cpuIDs, int(i))
			}
		}
	}

	sort.Ints(cpuIDs)

	return cpuIDs
}
