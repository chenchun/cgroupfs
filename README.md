# cgroupfs

 [![Build Status](https://travis-ci.org/chenchun/cgroupfs.svg?branch=master)](https://travis-ci.org/chenchun/cgroupfs)

Like lxcfs https://github.com/lxc/lxcfs, cgroupfs provides an emulated /proc/meminfo, /proc/cpuinfo... for containers.

# build
    make

# Usage

    container_id=`docker create -v /tmp/cgroupfs/meminfo:/proc/meminfo -m=15m ubuntu sleep 213133`

    ## In the second console tab
    mkdir /tmp/cgroupfs
    ./cgroupfs /tmp/cgroupfs /docker/$container_id

    ## Go to the first tab
    docker start $container_id

    ## Take a look at /tmp/cgroupfs/meminfo now
    ## cgroupfs file system should be able to show the memory usage of the container
    root@linux-dev:/home/vagrant# cat /tmp/cgroupfs/meminfo
    MemTotal:       15360 kB
    MemFree:        13432 kB
    MemAvailable:   13432 kB
    Buffers:        0 kB
    Cached:         1804 kB
    SwapCached:     0 kB

    ## Enter docker container, you should see free is showing the real usage
    docker exec -it $container_id bash
    root@251d4d18bca6:/# free -m
                 total       used       free     shared    buffers     cached
    Mem:            15          2         12          0          0          1
    -/+ buffers/cache:          0         14
    Swap:            0          0          0

# FAQ


**1. fusermount: exec: "fusermount": executable file not found in $PATH**

You should install fuse


    On debian/ubuntu
    sudo apt-get install fuse

**2. meminfo file cannot be mounted because it is located inside "/proc"**

You should update docker to 1.11+ or patch the related changes https://github.com/opencontainers/runc/pull/452, https://github.com/opencontainers/runc/pull/560

See related issues https://github.com/opencontainers/runc/issues/400
