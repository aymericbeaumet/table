.PHONY: all
all:

.PHONY: dependencies
dependencies:
	brew install ansible go golangci/tap/golangci-lint watchexec

.PHONY: dev-api
dev-api:
	cd ./cmd/api && watchexec --restart --clear -w . -w ../../pkg -- go run -race .

.PHONY: dev-outpost
dev-outpost:
	cd ./cmd/outpost && watchexec --restart --clear -w . -w ../../pkg -- go run -race .

.PHONY: dev-www
dev-www:
	cd ./www && npm run dev

.PHONY: update
update:
	cd ./cmd/api && go get -u
	cd ./cmd/outpost && go get -u
	go mod tidy

.PHONY: ci
ci:
	go mod download

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -race -v -failfast -vet=off ./...

.PHONY: lint
lint:
	gofmt -l .
	golangci-lint run ./...

.PHONY: deploy
deploy:
	ansible-playbook ./ansible/playbook.yml -i ./ansible/hosts

.PHONY: format
format:
	gofmt -s -w .
