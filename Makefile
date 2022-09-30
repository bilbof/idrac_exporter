.PHONY: test
build:
	go build -o bin/ .

# Builds in a golang docker container, no need for go locally
.PHONY: build-in-docker
build-in-docker:
	docker run -t -v $$PWD:/go/src/github.com/marshallwace/idrac_exporter -w /go/src/github.com/marshallwace/idrac_exporter golang:1.18 make build

.PHONY: run
run:
	./bin/idrac_exporter -config example.config.yml

.PHONY: docker-build
docker-build:
	docker build -t idrac_exporter .

.PHONY: docker-run
docker-run:
	docker run -it -v $$PWD:/etc/idrac -p 9348:9348 idrac_exporter:latest -config /etc/idrac/example.config.yml
