GOBIN = $(CURDIR)/bin/main
GOBUILD = env GO111MODULE=on go build 

build:
	$(GOBUILD) -o $(GOBIN) ./...


# env GO111MODULE=on go clean -cache
clean:
	rm -rf $(CURDIR)/bin
	rm *.dot *.svg