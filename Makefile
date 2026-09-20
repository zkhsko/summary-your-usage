.DEFAULT_GOAL := help
CONFIG ?= config.json

.PHONY: help run ui-install ui-dev ui-build build test vet fmt migrate-up migrate-down migrate-status

help:
	@printf '%s\n' \
	  'make run             启动后端（自动执行 goose up）' \
	  'make ui-install      安装前端锁定依赖' \
	  'make ui-dev          启动前端开发服务' \
	  'make build           构建前端与 Go 可执行文件' \
	  'make test            运行 Go 测试（含竞态检查）' \
	  'make vet             运行 Go 静态检查' \
	  'make migrate-status  查看迁移状态' \
	  'make migrate-up      执行所有待运行迁移' \
	  'make migrate-down    回滚最近一条迁移'

run: ui-build
	go run ./cmd/app -config "$(CONFIG)"

ui-install:
	npm --prefix ui ci

ui-dev:
	npm --prefix ui run dev

ui-build:
	npm --prefix ui run build

build: ui-build
	mkdir -p bin
	go build -o bin/app ./cmd/app
	go build -o bin/migrate ./cmd/migrate

test: ui-build
	go test -race ./...

vet: ui-build
	go vet ./...

fmt:
	gofmt -w cmd config internal pkg migrations ui/embed.go

migrate-up:
	go run ./cmd/migrate -config "$(CONFIG)" up

migrate-down:
	go run ./cmd/migrate -config "$(CONFIG)" down

migrate-status:
	go run ./cmd/migrate -config "$(CONFIG)" status
