.PHONY: build

test:
	@go test -cover -race ./... | grep -v '^?'
fmt:
	go fmt ./...
count-lines:
	@echo "total code lines:" && find . -name "*.go" -exec cat {} \; | wc -l

build:
	@echo "⚙️  Installing web player dependencies..."
	@npm ci --prefix ./web/player
	
	@echo "⚙️  Installing web studio dependencies..."
	@npm ci --prefix ./web/studio
	
	@echo "🛠️  Building web player..."
	@npm run build --prefix ./web/player
	
	@echo "🛠️  Building web studio..."
	@npm run build --prefix ./web/studio
	
	@echo "🛠️ Building web server..."
	@go build ./cmd/main.go
	
	@echo "✅ Build completed successfully"