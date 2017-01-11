Name := zabbixcli
Version := $(shell git describe --tags --abbrev=0)
OWNER := youyo
.DEFAULT_GOAL := help

## Setup
setup:
	go get github.com/kardianos/govendor
	go get github.com/Songmu/make2help/cmd/make2help

## Install dependencies
deps: setup
	govendor sync

## Initialize and Update dependencies
update: setup
	rm -rf /vendor/vendor.json
	govendor fetch +outside

## Vet
vet: setup
	govendor vet +local

## Lint
lint: setup
	go get github.com/golang/lint/golint
	govendor vet +local
	for pkg in $$(govendor list -p -no-status +local); do \
		golint -set_exit_status $$pkg || exit $$?; \
	done

## Run tests
test: deps
	govendor test +local -cover

## Run mysql-server
start-mysql:
	docker-compose up -d

## Destroy mysql-server
stop-mysql:
	docker-compose stop
	docker-compose rm -f

## Show help
help:
	@make2help $(MAKEFILE_LIST)

.PHONY: setup deps update vet lint test zabbix-build zabbix-destroy help
