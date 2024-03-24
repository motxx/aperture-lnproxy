all: build
build:
	go build ./cmd/app
	go build ./cmd/appcli
build-cli:
	go build ./cmd/appcli
install:
	go install ./cmd/app
	go install ./cmd/appcli
rpc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative contentrpc/content.proto
rpc-format:
	cd ./contentrpc; find . -name "*.proto" | xargs clang-format --style=file -i
