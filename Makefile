.PHONY: dev build install

dev:
	@cd web && npm run dev:all

build:
	@cd web && npm run build
	@go build -o commito .

install:
	@cd web && npm run build
	@go install .
