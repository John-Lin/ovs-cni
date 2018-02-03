all: bin

bin: clean
	sh build.sh

clean:
	rm -rf bin
docker-build: bin
	docker build -t hwchiu/ovs-cni:latest -f Docker/Dockerfile .

docker-push: docker-build
	docker push hwchiu/ovs-cni:latest
test:
	go get -u github.com/pierrre/gotestcover
	sudo -E env PATH=$$PATH TEST_ETCD=1 gotestcover -coverprofile=coverage.txt -covermode=atomic ./...
