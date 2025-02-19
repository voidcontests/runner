TARGETS = runner_linux_amd64 runner

all: $(TARGETS)

runner: main.go
	go build -o runner main.go

runner_linux_amd64: main.go
	CGO_ENABlED=0 GOOS=linux GOARCH=amd64 go build -o runner_linux_amd64 main.go

clean:
	rm -f $(TARGETS)

.PHONY: all clean
