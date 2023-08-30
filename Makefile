.EXPORT_ALL_VARIABLES:
GO111MODULE = on
TESTABLE_PACKAGES = `go list ./... | egrep -v 'testing|constants|cmd' | grep 'pitaya-bot/'`

setup:
	# NOOP

setup-ci:
	@go get github.com/mattn/goveralls
	@go get -u github.com/wadey/gocovmerge

setup-protobuf-macos:
	@brew install protobuf
	@go install github.com/gogo/protobuf/protoc-gen-gogofaster

run-testing-json-server:
	@docker-compose -f ./testing/json/docker-compose.yml up -d etcd nats redis && go run ./testing/json/main.go

run-testing-json-bots:
	@go run *.go run --duration 5s -d ./testing/json/specs/ --config ./testing/json/config/config.yaml

kill-testing-json-deps:
	@docker-compose -f ./testing/json/docker-compose.yml down; true

run-testing-proto-server:
	@(cd ./testing/protobuffer/ && make protos-compile)
	@docker-compose -f ./testing/protobuffer/docker-compose.yml up -d etcd nats && go run ./testing/protobuffer/main.go

run-testing-proto-bots:
	@go run *.go run --duration 5s -d ./testing/protobuffer/specs/ --config ./testing/protobuffer/config/config.yaml

kill-testing-proto-deps:
	@docker-compose -f ./testing/protobuffer/docker-compose.yml down; true

build-mac:
	@mkdir -p out
	@GOOS=darwin GOARCH=amd64 go build -o ./out/pitaya-bot-darwin ./main.go

build-linux:
	@mkdir -p out
	@GOOS=linux GOARCH=amd64 go build -o ./out/pitaya-bot-linux ./main.go

test: unit-test-coverage

unit-test-coverage:
	@echo "===============RUNNING UNIT TESTS==============="
	@GO111MODULE=on go test $(TESTABLE_PACKAGES) -coverprofile coverprofile.out

build-docker-image: build-linux
	@docker build -t pitaya-bot . -f Dockerfile-dev


