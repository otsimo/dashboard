.PHONY: default build docker release test clean gcr gcrpush

default: build

build: clean vet
	script/build none none

docker: clean vet
	script/build docker package

release: clean vet
	script/build docker
	script/docker
    
gcr: clean vet
	script/build docker package
	script/docker none gcr

gcrpush: clean vet
	script/build docker package
	script/docker push gcr

fmt:
	goimports -w src

vet:
	go vet ./src/...

test:
	script/test

clean:
	rm -rf bin
