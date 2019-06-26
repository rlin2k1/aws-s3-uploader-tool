GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=main

all: run

build: 
		$(GOBUILD) ./src/$(BINARY_NAME).go

clean: 
		$(GOCLEAN)
		rm -f ./src/$(BINARY_NAME)

run:
		$(GOBUILD) ./src/$(BINARY_NAME).go
		./$(BINARY_NAME)