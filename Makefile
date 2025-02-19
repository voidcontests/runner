TARGETS = bin/runner_linux_amd64 bin/runner
COMMAND = gcc /sandbox/main.c ; cat /sandbox/input.txt | timeout 5s /sandbox/a.out ; find /sandbox -type f -name "a.out" -delete

all: $(TARGETS)

bin/runner: main.go
	go build -o bin/runner main.go

bin/runner_linux_amd64: main.go
	CGO_ENABlED=0 GOOS=linux GOARCH=amd64 go build -o bin/runner_linux_amd64 main.go

run:
	docker run --rm --cpus=0.5 --memory=128m --memory-swap=256m --pids-limit=50 --read-only --security-opt seccomp=seccomp.json --network=none -v $(pwd)/example:/sandbox runner bash -c '$(COMMAND)'

clean:
	rm -f $(TARGETS)

.PHONY: all clean
