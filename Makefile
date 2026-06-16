.PHONY: build install uninstall ontime launchd logs clean

GO ?= go
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
REPORTS_DIR := $(HOME)/git/timeon/reports

build:
	$(GO) build -o bin/timeon ./cmd/timeon
	$(GO) build -o bin/ontime ./cmd/ontime

install: build
	install -d $(BINDIR)
	install -m 755 bin/timeon $(BINDIR)/timeon
	install -m 755 bin/ontime $(BINDIR)/ontime
	install -d $(REPORTS_DIR)

launchd: install
	sed "s|HOME_PLACEHOLDER|$(HOME)|g" launchd/com.duhd.timeon.plist > ~/Library/LaunchAgents/com.duhd.timeon.plist
	launchctl bootout gui/$$(id -u) ~/Library/LaunchAgents/com.duhd.timeon.plist 2>/dev/null || true
	launchctl bootstrap gui/$$(id -u) ~/Library/LaunchAgents/com.duhd.timeon.plist
	launchctl enable gui/$$(id -u)/com.duhd.timeon
	@echo "timeon installed and started via launchd"

uninstall:
	launchctl bootout gui/$$(id -u) ~/Library/LaunchAgents/com.duhd.timeon.plist 2>/dev/null || true
	rm -f ~/Library/LaunchAgents/com.duhd.timeon.plist
	rm -f $(BINDIR)/timeon $(BINDIR)/ontime

logs:
	tail -f $(REPORTS_DIR)/timeon.log $(REPORTS_DIR)/timeon.err

clean:
	rm -rf bin/
