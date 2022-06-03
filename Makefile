CGO_ENABLED := 0
export CGO_ENABLED

files := $(shell find . -name '*.go' -print)
pkgs := $(shell go list ./...)

moduleDir := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
build_dir := $(moduleDir)/build
dist_dir := $(moduleDir)/dist

# git commands won't work if there is no .git folder in the module or in the parent module
git_rev := $(shell git rev-parse --short HEAD)
git_tag := $(shell git tag --points-at=$(git_rev) | egrep -o "[0-9]+\.[0-9]+\.[0-9]+")
version := $(if $(git_tag),v$(git_tag),dev-$(git_rev))
release_version=$(if $(releaseVersion),$(releaseVersion),$(version))
build_time := $(shell date -u)
ldflags := -X "github.com/supreethrao/automated-rota-manager/cmd.version=$(release_version)" -X "github.com/supreethrao/automated-rota-manager/cmd.buildTime=$(build_time)"

# Define cross compiling targets
os := $(shell uname)
ifeq ("$(os)", "Linux")
	target_os = linux
	cross_os = darwin
else ifeq ("$(os)", "Darwin")
	target_os = darwin
	cross_os = linux
endif

PLATFORMS := linux darwin windows

# ------ TARGETS -------

.PHONY: setup
setup:
	@echo "== setup"
	go get -v golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports

.PHONY: check
check : check-os

.PHONY: check-os
check-os:
ifndef target_os
	$(error Unsupported platform: ${os})
endif

#.PHONY: lint
#lint:
#	golangci-lint run --fix
#
.PHONY: test
test :
	@echo "== run tests"
	go test -v $(pkgs)

.PHONY: build
build : check test
	@echo "== build"
	GOOS=${target_os} GOARCH=amd64 go build -ldflags '-s $(ldflags)' -o ${build_dir}/${target_os}/automated-rota-manager -v

.PHONY: build-all-platforms
build-all-platforms : check test
	@echo "== building binary for linux amd64"
	GOOS=${target_os} GOARCH=amd64 go build -ldflags '-s $(ldflags)' -o ${build_dir}/${target_os}/automated-rota-manager -v
	@echo "Cross compiling and building binary for ${cross_os} amd64"
	GOOS=${cross_os} GOARCH=amd64 go build -ldflags '-s $(ldflags)' -o ${build_dir}/${cross_os}/automated-rota-manager -v
	@echo "Cross compiling binary for windows amd64"
	GOOS=windows GOARCH=amd64 go build -ldflags '-s $(ldflags)' -o ${build_dir}/windows/automated-rota-manager.exe -v

.PHONY: install
install : build-all-platforms
	@echo "== installing binaries to dist folder"
	@mkdir -p $(dist_dir)
	@for platform in $(PLATFORMS); \
	do \
		echo "Installing $${platform} binary"; \
		cp -r $(build_dir)/$${platform} $(dist_dir)/$${platform}; \
	done;

.PHONY: docker-local
docker-local: build
	@echo "docker"
	@echo "$(version)"
	@cd ${moduleDir}
	@docker build -t local/automated-rota-manager:v$(version) .

.PHONY: clean
clean :
	@echo "== clean"
	rm -rf build
