OUTDIR	?= ./.out

.PHONY: all
all: lint test

.PHONY: lint
lint: _lint gci

.PHONY: gci
gci:
	GCIMODULE=`go list -m` envsubst < .golangci.gcitpl.yml | golangci-lint run -c /dev/stdin $(LINTARGS) $(LINTPATH)

.PHONY: _lint
_lint:
	golangci-lint run -c .golangci.yml $(LINTARGS) $(LINTPATH)

.PHONY: test
test: _OUTDIR
	go test -coverprofile="$(OUTDIR)/cover.out" ./... && go tool cover -func="$(OUTDIR)/cover.out"

.PHONY: cover
cover: test
	go tool cover -html=$(OUTDIR)/cover.out

.PHONY: clean
clean:
	rm -rf "$(OUTDIR)"

.PHONY: _OUTDIR
_OUTDIR:
	mkdir -p "$(OUTDIR)"
