OUTDIR	?= ./.out

all: lint test

lint: lint-main lint-gci

lint-main:
	golangci-lint run -c .golangci.yml $(LINTARGS) $(LINTPATH)

lint-gci:
	GCIMODULE=`go list -m` envsubst < .golangci.gcitpl.yml | golangci-lint run -c /dev/stdin $(LINTARGS) $(LINTPATH)

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
	lint
	lint-main
	lint-gci
	test
	cover
	clean
	_OUTDIR
