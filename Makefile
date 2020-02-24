GOPATH := $(shell pwd)
GOBIN := $(GOPATH)/bin
GO := go

proxy:
	cd $(GOPATH)/src/proxy/main && $(GO) build -o $(GOBIN)/proxy main.go

dbase:
	cd $(GOPATH)/src/db/main && $(GO) build -o $(GOBIN)/dbase main.go

image:
	cd $(GOPATH)/src/proxy/main && docker build -t proxy .

test:
	cd $(GOPATH)/src/proxy && $(GO) test -v -cover 
	cd $(GOPATH)/src/db && $(GO) test -v -cover

fmt:
	cd $(GOPATH)/src/proxy && $(GO) fmt
	cd $(GOPATH)/src/proxy/main && $(GO) fmt
	cd $(GOPATH)/src/db/main && $(GO) fmt

clean:
	rm -f $(GOPATH)/bin/proxy
	rm -f $(GOPATH)/bin/dbase

release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
	goreleaser --rm-dist

.PHONY: all

all: proxy dbase fmt test
 
