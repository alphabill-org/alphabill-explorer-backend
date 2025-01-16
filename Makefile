all: clean tools test build swagger

clean:
	rm -rf build/

test:
	go test ./... -coverpkg=./... -count=1 -coverprofile test-coverage.out

build:
	cd ./cmd && go build -tags=viper_bind_struct -o ../build/abexplorer

swagger:
	swag init --generalInfo restapi/routes.go --parseInternal --parseDependency --output restapi/docs

tools:
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: \
	all \
	clean \
	test \
	build \
	swagger \
	tools
