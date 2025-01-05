test:
	@go test -cover ./... | grep -v '^?'
fmt:
	go fmt ./...
count-lines:
	@echo "total code lines:" && find . -name "*.go" -exec cat {} \; | wc -l