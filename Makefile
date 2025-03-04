all: clean tools test build swagger

clean:
	rm -rf build/

test:
	go test ./... -coverpkg=./... -count=1 -coverprofile test-coverage.out

build:
	cd ./cmd && go build -tags=viper_bind_struct -o ../build/abexplorer

swagger:
	swag init --generalInfo api/routes.go --parseInternal --parseDependency --output api/docs

tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/vektra/mockery/v2@latest

generate-mocks:
	mockery

generate-mocks:
	mockery

.PHONY: \
	all \
	clean \
	test \
	build \
	swagger \
	tools \
	generate-mocks
