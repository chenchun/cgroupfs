build_image=cgroupfsbuild
container_name=cgroupfs
binary_name=${container_name}

all:docker-build get-binary

docker-build:
	@docker build -t ${build_image} .

get-binary:
	@echo "Copy binary \"cgroupfs\" from image to current directory..."
	@docker run --name=${container_name} -d ${build_image} sleep 1000
	@docker cp ${container_name}:/tmp/${binary_name} .
	@docker rm -v -f ${container_name}

test:docker-build
	@docker run --rm --privileged ${build_image} go test -v github.com/chenchun/cgroupfs

clean:
	rm -f ${binary_name}
	@docker rmi ${build_image}