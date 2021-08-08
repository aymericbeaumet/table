MODE ?= production
BIN := outpost-$(MODE)

.PHONY: all
all:

.PHONY: dependencies
dependencies:
	brew install go golangci/tap/golangci-lint watchexec

.PHONY: ci
ci:
	go mod download

.PHONY: update
update:
	go get -u && go mod tidy

.PHONY: dev
dev:
	watchexec --restart --clear -e go -- go run -race .

.PHONY: build
build:
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN) .

.PHONY: preview
preview:
	./$(BIN)

.PHONY: test
test:
	go test -v -failfast -vet=off -race ./...

.PHONY: lint
lint:
	out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "$$out"; exit 1; fi
	golangci-lint run ./...

.PHONY: deploy
deploy: outpost-0

.PHONY: outpost-%
outpost-%: build
	scp $(BIN) .env.$(MODE) $(MODE).{crt,key} cloudflare.crt $@:/usr/local/bin/
	scp outpost.service $@:/etc/systemd/system/
	ssh $@ " \
		systemctl enable outpost.service; \
		systemctl stop outpost.service; \
		mv /usr/local/bin/$(BIN) /usr/local/bin/outpost; \
		systemctl restart outpost.service; \
	"

.PHONY: format
format:
	gofmt -s -w .
