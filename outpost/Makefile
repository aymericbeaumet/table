BIN := outpost

.PHONY: all
all: build test lint

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
	true

.PHONY: lint
lint:
	true

.PHONY: format
format:
	true

.PHONY: deploy
deploy: table-outpost-0

.PHONY: table-outpost-%
table-outpost-%:
	scp $(BIN) $@:/usr/local/bin/$(BIN)-next
	scp .env.production $@:/usr/local/bin/
	scp outpost.service $@:/etc/systemd/system/
	ssh $@ "systemctl enable outpost.service"
	ssh $@ "systemctl stop outpost.service"
	ssh $@ "mv /usr/local/bin/$(BIN)-next /usr/local/bin/$(BIN)"
	ssh $@ "systemctl start outpost.service"
