.PHONY: serve
serve: lint
	go build -o asta && ./asta serve

.PHONY: lint
lint:
	go mod tidy
	golangci-lint run --fix ./...

.PHONY: docker-compose-up
docker-compose-up:
	docker-compose -f hack/docker-compose.yaml up -d

.PHONY: docker-compose-down
docker-compose-down:
	docker-compose -f hack/docker-compose.yaml down
