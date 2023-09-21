# Packages version to release
VERSION := 1.5.0

DIST_DIR := ../dist

LINUX_AMD64_BIN=$(DIST_DIR)/lefthook_linux_amd64_v1/lefthook
LINUX_ARM64_BIN=$(DIST_DIR)/lefthook_linux_arm64/lefthook
FREEBSD_AMD64_BIN=$(DIST_DIR)/lefthook_freebsd_amd64_v1/lefthook
FREEBSD_ARM64_BIN=$(DIST_DIR)/lefthook_freebsd_arm64/lefthook
WINDOWS_AMD64_BIN=$(DIST_DIR)/lefthook_windows_amd64_v1/lefthook.exe
WINDOWS_ARM64_BIN=$(DIST_DIR)/lefthook_windows_arm64/lefthook.exe
DARWIN_AMD64_BIN=$(DIST_DIR)/lefthook_darwin_amd64_v1/lefthook
DARWIN_ARM64_BIN=$(DIST_DIR)/lefthook_darwin_arm64/lefthook

prepare: clean set-version put-readme put-binaries

publish:
	cd npm; find . -type d -name 'lefthook*' -exec npm publish --access public \{} \; ; cd -
	cd npm-bundled; npm publish --access public ; cd -
	cd npm-installer; npm publish --access public ; cd -
	cd rubygems; rake build; gem push pkg/*.gem ; cd -

# Update versions of all packages
set-version:
	find npm -name 'package.json' -type f -print0 | xargs -0 sed -E -i "s/\"version\": \".+\"/\"version\": \"$(VERSION)\"/"
	sed -E -i "s/\"(lefthook-.+)\": \".+\"/\"\1\": \"$(VERSION)\"/g" npm/lefthook/package.json
	sed -E -i "0,/version/{s/\"version\": \".+\"/\"version\": \"$(VERSION)\"/}" npm-bundled/package.json
	sed -E -i "0,/version/{s/\"version\": \".+\"/\"version\": \"$(VERSION)\"/}" npm-installer/package.json
	sed -E -i "s/(spec\.version\s+= ).*/\1\"$(VERSION)\"/" rubygems/lefthook.gemspec

put-binaries:
	install -D $(LINUX_AMD64_BIN) npm/lefthook-linux-x64/bin/lefthook
	install -D $(LINUX_ARM64_BIN) npm/lefthook-linux-arm64/bin/lefthook
	install -D $(FREEBSD_AMD64_BIN) npm/lefthook-freebsd-x64/bin/lefthook
	install -D $(FREEBSD_ARM64_BIN) npm/lefthook-freebsd-arm64/bin/lefthook
	install -D $(WINDOWS_AMD64_BIN) npm/lefthook-windows-x64/bin/lefthook.exe
	install -D $(WINDOWS_ARM64_BIN) npm/lefthook-windows-arm64/bin/lefthook.exe
	install -D $(DARWIN_AMD64_BIN) npm/lefthook-darwin-x64/bin/lefthook
	install -D $(DARWIN_ARM64_BIN) npm/lefthook-darwin-arm64/bin/lefthook
	cd npm-bundled; npm version $(VERSION) --allow-same-version; cd -
	cd $(DIST_DIR); find . -maxdepth 2 -type f -exec cp --parents \{\} ../packaging/rubygems/libexec/ \;

put-readme:
	find npm/ -type d -name 'lefthook*' -exec cp -f ../README.md \{} \;
	cp ../README.md npm-bundled/
	cp ../README.md npm-installer/

clean:
	find npm/ -name 'README.md' -exec rm \{} \;
	find npm/ -type f -name 'lefthook*' -exec rm \{} \;
	git clean -fdX npm-installer/ npm-bundled/ npm-bundled/bin/
	git clean -fdX rubygems/libexec/ rubygems/pkg/
