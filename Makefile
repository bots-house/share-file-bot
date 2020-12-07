sqlboiler_version = 4.3.1
sqlboiler_download_url = https://api.github.com/repos/volatiletech/sqlboiler/tarball/v$(sqlboiler_version)

golangci_lint_version = 1.31.0

run: services
	go run main.go -config .env.local

lint: .bin/golangci-lint
	.bin/golangci-lint run --config .golangci.yml

generate:  generate-dal

generate-dal: .bin/sqlboiler .bin/sqlboiler-psql
	.bin/sqlboiler .bin/sqlboiler-psql

generate-domain: core/kind_string.go core/chattype_string.go bot/state/state_string.go

services:
	docker-compose up --no-recreate --detach postgres redis 

core/kind_string.go: core/kind.go
	cd core && stringer -type Kind -trimprefix Kind

core/chattype_string.go: core/chat.go
	cd core && stringer -type ChatType -trimprefix ChatType

bot/state/state_string.go: bot/state/state.go
	cd bot/state && stringer -type State

.bin/golangci-lint:
	mkdir -p .bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b .bin v$(golangci_lint_version)

.bin/sqlboiler .bin/sqlboiler-psql:
	mkdir -p .bin
	curl -o .bin/sqlboiler.tar.gz -L $(sqlboiler_download_url)
	tar -xzf .bin/sqlboiler.tar.gz --directory .bin && rm .bin/sqlboiler.tar.gz
	mv .bin/volatiletech-sqlboiler-* .bin/sqlboiler-src
	cd .bin/sqlboiler-src && go build -o ../sqlboiler
	cd .bin/sqlboiler-src/drivers/sqlboiler-psql && go build -o ${CURDIR}/.bin/sqlboiler-psql
	rm -r .bin/sqlboiler-src

psql:
	docker-compose exec postgres psql -U sfb

redis-cli:
	docker-compose exec redis redis-cli

psql-recreate-db:
	docker-compose exec postgres dropdb --username sfb sfb
	docker-compose exec postgres createdb --username sfb sfb 