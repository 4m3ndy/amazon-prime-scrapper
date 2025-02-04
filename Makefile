IMAGE_REGISTRY_URI?=docker.io/4m3ndy/amazon-scrapper-service
IMAGE_TAG?=v1.0.0
GOLANG_IMAGE?=golang:1.15.1-alpine3.12

init:
	go mod download

build:
	docker run -it --rm -v `pwd`:/app -w /app -e GO111MODULE=on -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 \
	${GOLANG_IMAGE} sh -c "go build -v -x -o /app/bin/main /app/cmd/server/*.go"

run:
	AMAZON_SCRAPPER_SVC_HTTP_PORT=8080 go run cmd/server/main.go

docker-build:
	docker build . -f Dockerfile -t ${IMAGE_REGISTRY_URI}:${IMAGE_TAG}

docker-push:
	docker push ${IMAGE_REGISTRY_URI}:${IMAGE_TAG}
