build_image=cgroupfsbuild
container_name=cgroupfs
binary_name=${container_name}

all:docker-build get-binary

docker-build:
	@docker build -t ${build_image} .

get-binary:
	@echo "copy binary from image..."
	@docker run --name=${container_name} -d ${build_image} sleep 1000
	@docker cp ${container_name}:/tmp/${binary_name} .
	@docker rm -v -f ${container_name}

clean:
	rm -f ${binary_name}
	@docker rmi ${build_image}