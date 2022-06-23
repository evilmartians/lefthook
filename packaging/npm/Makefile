# NPM package version to release
VERSION := 1.0.0

DIST_DIR := ../../dist

LINUX_AMD64_BIN=$(DIST_DIR)/lefthook_linux_amd64_v1/lefthook
LINUX_ARM64_BIN=$(DIST_DIR)/lefthook_linux_arm64/lefthook
WINDOWS_AMD64_BIN=$(DIST_DIR)/lefthook_windows_amd64_v1/lefthook.exe
WINDOWS_ARM64_BIN=$(DIST_DIR)/lefthook_windows_arm64/lefthook.exe
DARWIN_AMD64_BIN=$(DIST_DIR)/lefthook_darwin_amd64_v1/lefthook
DARWIN_ARM64_BIN=$(DIST_DIR)/lefthook_darwin_arm64/lefthook

prepare: clean set-version put-readme put-binaries

publish:
	find . -type d -name 'lefthook*' -exec npm publish --access public \{} \;

# Update versions of all packages
set-version:
	find . -name 'package.json' -type f -print0 | xargs -0 sed -E -i "s/\"version\": \".+\"/\"version\": \"$(VERSION)\"/"
	sed -E -i "s/\"(lefthook-.+)\": \".+\"/\"\1\": \"$(VERSION)\"/g" lefthook/package.json

put-binaries:
	install -D $(LINUX_AMD64_BIN) lefthook-linux-x64/bin/lefthook
	install -D $(LINUX_ARM64_BIN) lefthook-linux-arm64/bin/lefthook
	install -D $(WINDOWS_AMD64_BIN) lefthook-windows-x64/bin/lefthook.exe
	install -D $(WINDOWS_ARM64_BIN) lefthook-windows-arm64/bin/lefthook.exe
	install -D $(DARWIN_AMD64_BIN) lefthook-darwin-x64/bin/lefthook
	install -D $(DARWIN_ARM64_BIN) lefthook-darwin-arm64/bin/lefthook

put-readme:
	find . -type d -name 'lefthook*' -exec cp -f ../../README.md \{} \;

clean:
	find . -name 'README.md' -exec rm \{} \;
	find . -type f -executable -exec rm \{} \;
