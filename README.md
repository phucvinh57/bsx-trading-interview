# BSX trading challenge

## Prerequisites

- Install libc: `sudo apt install librocksdb-dev libsnappy-dev libz-dev liblz4-dev libzstd-dev`

## Tasks:

- [x] Refactor code
- [x] Get user's orders from the order book
- [x] Write tests
- [ ] Write benchmarks
- [ ] Write documentation

## Testing

To make environment variables available to the test runner, add this line to VSCode `setting.json`

```json
"go.testEnvFile": "${workspaceFolder}/.env"
```