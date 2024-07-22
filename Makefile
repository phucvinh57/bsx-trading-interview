.PHONY: build

build:
	CGO_CFLAGS="-I/usr/include/rocksdb" \
	CGO_LDFLAGS="-L/usr/include/rocksdb -lrocksdb -lstdc++ -lm -lz -lsnappy -llz4 -lzstd" \
	go build cmd/main.go
