.PHONY: build test dev clean bench

dev: setup
	air -c .air.toml
build:
	CGO_CFLAGS="-I/usr/include/rocksdb" \
	CGO_LDFLAGS="-L/usr/include/rocksdb -lrocksdb -lstdc++ -lm -lz -lsnappy -llz4 -lzstd" \
	go build cmd/main.go
test: setup
	go test -v ./...
bench: setup
	go test -bench=. -benchmem -benchtime=10s ./test/bench
clean:
	rm -f main
	rm -rf tmp **/*/rocksdb_data rocksdb_data
	docker compose down --volumes
setup:
	docker compose up -d
	until (echo 'db.runCommand("ping").ok' | docker compose exec -T mongo mongosh --quiet); \
		do sleep 1; \
	done;