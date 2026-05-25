BIN := dist/yasp
CMD := ./cmd/yasp

GOBUILD := go build -trimpath -v
GOTEST := go test -cover -race

.PHONY: $(BIN)
$(BIN):
	$(GOBUILD) -o $@ $(CMD)

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: lint
lint: vet fix vuln

.PHONY: vuln
vuln:
	go tool govulncheck ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: fix
fix:
	go fix -diff ./...

.PHONY: fix-do
fix-do:
	go fix ./...
