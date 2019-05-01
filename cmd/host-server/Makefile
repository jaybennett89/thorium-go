# Inspired by Warren Van Winckel

pwd := $(shell pwd)

all: build

build:
	go build host-server.go

image: build
	docker build -t host-server .

run:
	docker run -it -d -v /opt/thorium/gameserver:/opt/thorium/gameserver --name=host-server host-server

stop:
	docker kill host-server

rm: stop
	docker rm host-server
