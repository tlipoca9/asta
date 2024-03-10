# asta

This repository contains a basic template that you can use via gonew.

```bash
go install golang.org/x/tools/cmd/gonew@latest
gonew github.com/tlipoca9/asta example.com/foo
sed -i 's/asta/foo/g' $(find . -type f | grep -v .git | grep -v README.md)
```
