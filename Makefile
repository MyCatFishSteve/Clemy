GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

UPXCMD=upx

BINARY_PATH=build
BINARY_NAME=clemy
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build-linux

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)/*

deps:
	$(GOGET) github.com/aws/aws-sdk-go/aws
	$(GOGET) github.com/aws/aws-sdk-go/aws/session
	$(GOGET) github.com/aws/aws-sdk-go/service/ec2
	$(GOGET) github.com/aws/aws-sdk-go/service/autoscaling

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_PATH)/$(BINARY_UNIX) -v
	$(UPXCMD) --ultra-brute $(BINARY_PATH)/$(BINARY_UNIX)
	zip $(BINARY_PATH)/$(BINARY_UNIX).zip $(BINARY_PATH)/$(BINARY_UNIX)
	rm -f $(BINARY_PATH)/*.upx $(BINARY_PATH)/$(BINARY_UNIX)

