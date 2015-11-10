# cgroupfs

Like lxcfs https://github.com/lxc/lxcfs, cgroupfs provides an emulated /proc/meminfo, /proc/cpuinfo... for the containers.

# Usage

    container_id=`docker create --device /dev/fuse --cap-add SYS_ADMIN -v /tmp/cgroupfs/meminfo:/proc/meminfo -m=15m ubuntu sleep 213133`

    ## in the second console tab
    go run cgroupfs.go /tmp/cgroupfs /docker/$container_id

    ## go to the first tab
    docker start $container_id

    docker exec -it $container_id bash
    root@251d4d18bca6:/# free -m
                 total       used       free     shared    buffers     cached
    Mem:            15          2         12          0          0          1
    -/+ buffers/cache:          0         14
    Swap:            0          0          0