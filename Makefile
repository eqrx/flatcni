export GITTAG=$(shell git describe --tags --always)
export GITCOMMIT=$(shell git log -1 --pretty=format:"%H")
export GOLDFLAGS=-s -w -extldflags '-zrelro -znow' -X go.eqrx.net/flatcni.version=$(GITTAG) -X go.eqrx.net/flatcni.commit=$(GITCOMMIT)
export GOFLAGS=-trimpath
export CGO_ENABLED=0

.PHONY: all
all: build

.PHONY: build
build: dist/arm64/flatcni dist/amd64/flatcni dist/arm/flatcni

.PHONY: dist/arm64/flatcni
dist/arm64/flatcni:
	GOARCH=arm64 go build -trimpath -ldflags "$(GOLDFLAGS)" -o $@ ./cmd/flatcni

.PHONY: dist/arm/flatcni
dist/arm/flatcni:
	GOARCH=arm go build -trimpath -ldflags "$(GOLDFLAGS)" -o $@ ./cmd/flatcni

.PHONY: dist/amd64/flatcni
dist/amd64/flatcni:
	go build -trimpath -ldflags "$(GOLDFLAGS)" -o $@ ./cmd/flatcni

.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem ./...

lint:
	golangci-lint run ./...

.PHONY: download
download:
	go mod download

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: update
update:
	go get -t -u=patch ./...
