# amazon-scrapper-service
Amazon Movie Scrapper based on Golang

## Prerequisites
- make >= 3.81
- docker >= 18.09.9
- kubectl >= 1.15

## How to run locally
``` bash
# If you want to use a different port, please define AMAZON_SCRAPPER_SVC_HTTP_PORT in the Makefile
# default value for AMAZON_SCRAPPER_SVC_HTTP_PORT is 8080
make init
make run
```



## How to build and deploy
``` bash
# make sure you define IMAGE_REGISTRY_URI and IMAGE_TAG
make build
make docker-build
make docker-push

# In order to deploy on kubernetes
kubectl apply -R -f ./k8s
```
