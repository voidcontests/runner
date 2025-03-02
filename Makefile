TARGETS = build/runner_linux_amd64 build/runner
COMMAND = gcc /sandbox/main.c ; cat /sandbox/input.txt | timeout 5s /sandbox/a.out ; find /sandbox -type f -name "a.out" -delete

all: $(TARGETS)

build/runner: main.go
	go build -o build/runner cmd/server/main.go

build/runner_linux_amd64: main.go
	CGO_ENABlED=0 GOOS=linux GOARCH=amd64 go build -o build/runner_linux_amd64 cmd/server/main.go

run:
	docker run --rm --cpus=0.5 --memory=128m --memory-swap=256m --pids-limit=50 --read-only --security-opt seccomp=seccomp.json --network=none -v $(pwd)/example:/sandbox runner bash -c '$(COMMAND)'

GIT_COMMIT := $(shell git rev-parse --short HEAD)

build:
	docker build -t jus1d/void-runner:latest -f ./docker/Dockerfile .

push: build
	docker tag jus1d/void-runner:latest jus1d/void-runner:$(GIT_COMMIT)
	@for tag in $(GIT_COMMIT) latest; do \
		docker push jus1d/void-runner:$$tag; \
	done

clean:
	rm -f $(TARGETS)

.PHONY: all clean build push run
