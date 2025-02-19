TARGETS = runner_linux_amd64 runner
COMMAND = gcc /sandbox/main.c ; cat /sandbox/input.txt | timeout 5s /sandbox/a.out ; find /sandbox -type f -name "a.out" -delete

all: $(TARGETS)

runner: main.go
	go build -o runner main.go

runner_linux_amd64: main.go
	CGO_ENABlED=0 GOOS=linux GOARCH=amd64 go build -o runner_linux_amd64 main.go

run:
	docker run --rm --cpus=0.5 --memory=128m --memory-swap=256m --pids-limit=50 --read-only --security-opt seccomp=seccomp.json --network=none -v $(pwd)/example:/sandbox runner bash -c '$(COMMAND)'

clean:
	rm -f $(TARGETS)

.PHONY: all clean
