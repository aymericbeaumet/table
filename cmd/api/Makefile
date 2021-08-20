NAME := api
SYSTEMCTL_SERVICE := $(NAME).service

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
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(NAME) .

.PHONY: test
test:
	go test -race -v -failfast -vet=off ./...

.PHONY: lint
lint:
	out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "$$out"; exit 1; fi
	golangci-lint run ./...

.PHONY: deploy
deploy: build
	# prepare service
	scp $(SYSTEMCTL_SERVICE) api.jointable.io:/etc/systemd/system/
	# prepare binary
	scp $(NAME) api.jointable.io:/usr/local/bin/$(NAME).next
	# prepare config
	ssh api.jointable.io "mkdir -p /usr/local/etc/$(NAME)"
	scp .env .env.production production.crt production.key cloudflare.crt private.pem public.pem api.jointable.io:/usr/local/etc/$(NAME)/
	# (re)start service
	ssh api.jointable.io " \
		systemctl enable $(SYSTEMCTL_SERVICE); \
		systemctl stop $(SYSTEMCTL_SERVICE); \
		mv /usr/local/bin/$(NAME).next /usr/local/bin/$(NAME); \
		systemctl restart $(SYSTEMCTL_SERVICE); \
	"

.PHONY: logs
logs:
	@ssh api.jointable.io "journalctl -f -o cat -u $(SYSTEMCTL_SERVICE)"

.PHONY: format
format:
	gofmt -s -w .
