# Inspired by Warren Van Winckel

pwd := $(shell pwd)

all: run

image:
	docker build -t thorium-db .

run: 
	docker run -it -d -p 54321:5432 -e POSTGRES_PASSWORD=secret -v ${pwd}:/docker-entrypoint-initdb.d --name=thorium-db library/postgres

stop:
	docker kill thorium-db
	docker rm thorium-db
