# Inspired by Warren Van Winckel

pwd := $(shell pwd)

all: build

build:
	cd cmd/host-server && make image
	cd cmd/masterserver && make image

run: 
	docker-compose up -d

stop:
	docker-compose stop || true

rm:
	docker-compose rm -f || true

restart: stop rm run
