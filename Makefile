GOCC=go
MKDIR_P=mkdir -p
GIT_SUBM=git submodule

BIN_PATH=./build
BIN="./build/block-scorer"

.PHONY: check dependencies build install clean

build: 
    $(GOCC) build -o $(BIN)


install:
    $(GOCC) install


dependencies:
    $(GIT_SUBM) update --init

clean:
    rm -r $(BIN_PATH)