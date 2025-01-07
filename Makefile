test:
	@go test -cover -race ./... | grep -v '^?'
fmt:
	go fmt ./...
count-lines:
	@echo "total code lines:" && find . -name "*.go" -exec cat {} \; | wc -l