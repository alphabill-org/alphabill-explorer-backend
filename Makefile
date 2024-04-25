all: clean tools test build gosec

clean:
	rm -rf build/

test:
	go test ./... -coverpkg=./... -count=1 -coverprofile test-coverage.out

build:
	cd ./cmd && go build -o ../build/abexplorer

.PHONY: \
	all \
	clean \
	test \
	build
