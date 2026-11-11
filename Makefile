GO ?= go

.PHONY: help install dev build test check clean
help:
	@printf 'install  下载依赖\ndev      原生开发模式\nbuild    构建本机桌面应用\ntest     Go 测试\ncheck    静态检查与编译验证\nclean    清理构建产物\n'
install:
	$(GO) mod download
dev:
	$(GO) tool mygo dev
build:
	CGO_ENABLED=0 $(GO) tool mygo build
test:
	CGO_ENABLED=0 $(GO) test ./...
check:
	$(GO) vet ./...
	$(GO) tool mygo vet .
	CGO_ENABLED=0 $(GO) test ./...
clean:
	$(GO) clean
	rm -rf build .mygo
