TESTABLE_PACKAGES = `go list ./... | egrep -v 'testing|constants|models|cmd' | grep 'pitaya-bot/'`

setup:
	@dep ensure

setup-ci:
	@go get github.com/mattn/goveralls
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/wadey/gocovmerge
	@dep ensure

setup-protobuf-macos:
	@brew install protobuf
	@go get -u github.com/gogo/protobuf/protoc-gen-gogofaster

run-testing-server:
	@docker-compose -f ./testing/docker-compose.yml up -d etcd nats && go run ./testing/main.go

run-testing-bots:
	@go run *.go run -d ./testing/specs/ --config ./testing/config/config.yaml

kill-testing-deps:
	@docker-compose -f ./testing/docker-compose.yml down; true

build-linux:
	@mkdir -p out
	@GOOS=linux GOARCH=amd64 go build -o ./out/pitaya-bot-linux ./main.go

test: unit-test-coverage

unit-test-coverage:
	@echo "===============RUNNING UNIT TESTS==============="
	@go test $(TESTABLE_PACKAGES) -coverprofile coverprofile.out


