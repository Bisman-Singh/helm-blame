BINARY = helm-blame
INSTALL_DIR = bin

.PHONY: build
build:
	@mkdir -p $(INSTALL_DIR)
	go build -o $(INSTALL_DIR)/$(BINARY) ./cmd/helm-blame

.PHONY: test
test:
	go test ./... -v

.PHONY: clean
clean:
	rm -rf $(INSTALL_DIR)
