TESTABLE_PACKAGES = `go list ./... | grep -v examples | grep -v constants | grep -v mocks | grep -v helpers | grep -v interfaces | grep -v protos | grep -v e2e | grep -v benchmark`

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

ensure-bin:
	@[ -f ./testing/server ] || go build -o ./testing/server ./testing/main.go

run-server:
	@go run ./testing/main.go

run-bots:
	@go run *.go run

ensure-deps:
	@cd ./testing && docker-compose up -d

kill-deps:
	@cd ./testing && docker-compose down; true

dev-deps: ensure-deps ensure-bin
