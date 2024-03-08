.PHONY: serve
serve: fmt
	go build -o asta && ./asta serve

.PHONY: fmt
fmt: tidy
	gofumpt -l -w .

.PHONY: tidy
tidy:
	go mod tidy
