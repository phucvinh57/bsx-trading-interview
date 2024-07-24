.PHONY: build

build:
	CGO_CFLAGS="-I/usr/include/rocksdb" \
	CGO_LDFLAGS="-L/usr/include/rocksdb -lrocksdb -lstdc++ -lm -lz -lsnappy -llz4 -lzstd" \
	go build cmd/main.go

test:
	docker compose up -d
	go test -v ./...
clean:
	rm -f main
	rm -rf tmp **/rocksdb_data rocksdb_data
	docker compose down --volumes