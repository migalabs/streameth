GOCC=go
MKDIR_P=mkdir -p
GIT_SUBM=git submodule

BIN_PATH=./build
BIN="./build/eth_cl_live_metrics"

.PHONY: check dependencies build install clean

build:
	$(GOCC) build -o $(BIN)

install:
	$(GOCC) install

dependencies:
	$(GIT_SUBM) update --init

clean:
	rm -r $(BIN_PATH)

run: 
	$(BIN) $(LIVE_METRICS_CMD) \
    	--log-level=${LIVE_METRICS_LOG_LEVEL} \
        --bn-endpoints=${LIVE_METRICS_BN_ENDPOINT} \
        --db-endpoint=${LIVE_METRICS_DB_URL} \
        --db-workers=${LIVE_METRICS_DB_WORKERS_NUM} \
		--metrics=${LIVE_METRICS_METRICS}