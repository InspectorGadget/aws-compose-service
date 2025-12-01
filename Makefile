BINARY_NAME := aws-compose-service
INSTALL_DIR := /usr/local/bin

all: build install clean

build:
	go build -o $(BINARY_NAME) .

install:
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
