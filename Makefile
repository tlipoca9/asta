.PHONY: serve
serve: fmt
	go build -o asta && ./asta serve

.PHONY: fmt
fmt: tidy
	goimports-reviser -set-alias -format -project-name github.com/tlipoca9/asta ./...
	gofumpt -l -w .

.PHONY: tidy
tidy:
	go mod tidy
