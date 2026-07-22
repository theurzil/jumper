BINARY := jumper
PREFIX := /usr/local/bin
SHELL_RC := $(HOME)/.bashrc
INSTALL_DIR := $(HOME)/.local/share/jumper
SCRIPT_DEST := $(INSTALL_DIR)/jumper.sh

.PHONY: build install uninstall clean test

build:
	go build -o $(BINARY) .

install: build
	install -Dm755 $(BINARY) $(PREFIX)/$(BINARY)
	install -Dm644 jumper.sh $(SCRIPT_DEST)
	@if ! grep -qs "source $(SCRIPT_DEST)" $(SHELL_RC); then \
		echo "source $(SCRIPT_DEST)" >> $(SHELL_RC); \
		echo "Added source line to $(SHELL_RC)"; \
	fi
	@echo "Installed. Run: source $(SHELL_RC)  (or restart your shell)"

uninstall:
	rm -f $(PREFIX)/$(BINARY)
	rm -rf $(INSTALL_DIR)
	@echo "Removed binary and $(INSTALL_DIR)."
	@echo "Manually remove the 'source $(SCRIPT_DEST)' line from $(SHELL_RC)."

test:
	go build -o /tmp/$(BINARY)-check .
	rm -f /tmp/$(BINARY)-check

clean:
	rm -f $(BINARY)
