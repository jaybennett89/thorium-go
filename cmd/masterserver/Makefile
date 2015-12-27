# Inspired by Warren Van Winckel

pwd := $(shell pwd)

all: build

build:
	go build master-server.go
	mv master-server ../../

image: build
	docker build -t master-server .

run: 
	docker run -it -d -v ${pwd}/../..:/usr/src/app --name=master-server master-server

stop:
	docker kill master-server

rm: stop
	docker rm master-server
