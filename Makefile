default: build

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" .

test:
	go test .
	golangci-lint run

bench:
	go test -bench=. -run=nothing

install:
	go install
