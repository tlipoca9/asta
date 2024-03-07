.PHONY: serve
serve: fmt
	go build -o asta && ./asta serve

.PHONY: fmt
fmt:
	gofumpt -l -w .