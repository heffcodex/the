OUTDIR	?= ./.out

all: fmt lint test

fmt:
	golangci-lint fmt -c .golangci.yml

lint:
	golangci-lint run -c .golangci.yml

test: _OUTDIR
	go test -coverprofile="$(OUTDIR)/cover.out" ./... && go tool cover -func="$(OUTDIR)/cover.out"

cover: test
	go tool cover -html=$(OUTDIR)/cover.out

clean:
	rm -rf "$(OUTDIR)"

_OUTDIR:
	mkdir -p "$(OUTDIR)"

.PHONY:
	all
	fmt
	lint
	test
	cover
	clean
	_OUTDIR
