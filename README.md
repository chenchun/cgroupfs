# cgroupfs

Like lxcfs https://github.com/lxc/lxcfs, cgroupfs provides an emulated /proc/meminfo, /proc/cpuinfo... for the containers.

# Usage

    go run cgroupfs.go /tmp/cgroupfs /docker/$container_id
    docker run --rm -it --device /dev/fuse --cap-add SYS_ADMIN -v /tmp/cgroupfs/meminfo:/proc/meminfo ubuntu bash

    root@91f2a72135cb:/# free -m
                 total       used       free     shared    buffers     cached
    Mem:            10          0          9          0          0          0
    -/+ buffers/cache:          0          9
    Swap:            0          0          0
