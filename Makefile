GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

SRC_FOLDER = src

BINARY_NAME=main

all: run

build: 
		$(GOBUILD) $(SRC_FOLDER)/$(BINARY_NAME).go

clean: 
		$(GOCLEAN)
		rm -f $(SRC_FOLDER)/$(BINARY_NAME)

run:
		$(GOBUILD) $(SRC_FOLDER)/$(BINARY_NAME).go
		./$(BINARY_NAME)