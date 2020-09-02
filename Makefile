init:
	go mod download

build:
	go build -v -x -o bin/main cmd/server/*.go

docker:
	docker build 