.PHONY: default build docker release test clean gcr gcrpush

gcrenv=prod

default: build

build: clean
	script/build none none

docker: clean
	script/build docker package

release: clean
	script/build docker
	script/docker
    
gcr: clean
	script/build docker package
	script/docker none gcr $(gcrenv)

gcrpush: clean
	script/build docker package
	script/docker push gcr $(gcrenv)

fmt:
	goimports -w src

vet:
	go vet ./src/...

test:
	script/test

clean:
	rm -rf bin
