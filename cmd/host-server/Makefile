# Inspired by Warren Van Winckel

pwd := $(shell pwd)

all: build

build:
	go build host-server.go

image: build
	docker build -t thorium/host-server .

run: stop rm
	docker run -d -v `pwd`/config:/opt/thorium/host-server/config -v /opt/thorium/gameserver:/opt/thorium/gameserver --name=host-server thorium/host-server

stop:
	docker kill host-server || true

rm:
	docker rm host-server || true
