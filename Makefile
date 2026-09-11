OUTDIR	?= $(PWD)/.out

all: go fmt lint test

go:
	@echo "go get & tidy:"
	@for dir in $$(go list -m -f '{{.Dir}}'); do \
  		echo "==> $$dir"; \
  		go -C "$$dir" get -u -t ./... || exit $$?; \
  		go -C "$$dir" mod tidy || exit $$?; \
  	done;

fmt:
	@echo "golangci-lint fmt:"
	@for dir in $$(go list -m -f '{{.Dir}}'); do \
		echo "==> $$dir"; \
		(cd "$$dir" && golangci-lint fmt -c "$(PWD)/.golangci.yml") || exit $$?; \
	done;

lint:
	@echo "golangci-lint run:"
	@for dir in $$(go list -m -f '{{.Dir}}'); do \
		echo "==> $$dir"; \
		(cd "$$dir" && golangci-lint run -c "$(PWD)/.golangci.yml") || exit $$?; \
	done;

test: _OUTDIR
	@echo "go test & coverprofile -> $(OUTDIR)/cover.out:"
	@rm -f "$(OUTDIR)"/*.cover.out "$(OUTDIR)/cover.out"; \
	first=1; \
	for dir in $$(go list -m -f '{{.Dir}}'); do \
		name=$$(basename "$$dir"); \
		echo "==> $$dir"; \
		(cd "$$dir" && go test -coverprofile="$(OUTDIR)/$$name.cover.out" ./...) || exit $$?; \
		if [ $$first -eq 1 ]; then \
			cat "$(OUTDIR)/$$name.cover.out" > "$(OUTDIR)/cover.out"; \
			first=0; \
		else \
			tail -n +2 "$(OUTDIR)/$$name.cover.out" >> "$(OUTDIR)/cover.out"; \
		fi; \
	done; \
	go tool cover -func="$(OUTDIR)/cover.out"

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
