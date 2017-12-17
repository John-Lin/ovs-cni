all: bin

bin:
	sh build.sh

clean:
	rm -rf bin
docker-build:
	docker build -t hwchiu/ovs-cni:latest -f Docker/Dockerfile .

docker-push: docker-build
	docker push hwchiu/ovs-cni:latest
