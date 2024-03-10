.PHONY: serve
serve: lint
	go build -o asta && ./asta serve

.PHONY: lint
lint: fmt
	golangci-lint run -v ./...

.PHONY: fmt
fmt: tidy
	goimports -l -w -local github.com/tlipoca9/asta .
	gofumpt -l -w .

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docker-compose-up
docker-compose-up:
	docker-compose -f hack/docker-compose.yaml up -d

.PHONY: docker-compose-down
docker-compose-down:
	docker-compose -f hack/docker-compose.yaml down
