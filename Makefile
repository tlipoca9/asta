.PHONY: serve
serve: lint
	go build -o run/asta && run/asta serve

.PHONY: lint
lint:
	go generate ./...
	go mod tidy
	golangci-lint run --fix ./...
	command -v deadcode 2>&1 > /dev/null || go install golang.org/x/tools/cmd/deadcode@latest
	deadcode ./...

.PHONY: docker-compose-up
docker-compose-up:
	docker-compose -f hack/docker-compose.yaml up -d

.PHONY: docker-compose-down
docker-compose-down:
	docker-compose -f hack/docker-compose.yaml down
