.PHONY: dev run build test race vet fmt check generate i18n snapshots icon icon-check

dev:
	go tool mygo dev

run:
	go run .

build: icon-check
	go tool mygo build

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

generate:
	go tool mygo generate

i18n:
	go tool mygo i18n check

check: test vet i18n icon-check

snapshots:
	NEURAL_BREACH_SNAPSHOTS="$(CURDIR)/snapshots" go test . -run TestViewSnapshots -count=1

icon:
	go run ./tools/genicon

icon-check:
	go run ./tools/genicon -check
