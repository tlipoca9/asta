# asta

This repository contains a basic template that you can use via gonew.

```bash
go install golang.org/x/tools/cmd/gonew@latest
gonew github.com/tlipoca9/asta example.com/foo
git init
git add .
git commit -m "gonew: github.com/tlipoca9/asta"
sed -i 's#github.com/tlipoca9/asta#example.com/foo#g' $(find . -type f | grep -v .git | grep -v README.md)
sed -i 's/asta/foo/g' $(find . -type f | grep -v .git | grep -v README.md)
```


## Quick Start

```bash
make docker-compose-up
make
```
