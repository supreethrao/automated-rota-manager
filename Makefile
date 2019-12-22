files := $(shell find . -name '*.go' -print)
pkgs := $(shell go list ./...)

build_dir := $(moduleDir)/build
dist_dir := $(moduleDir)/dist

# git commands won't work if there is no .git folder in the module or in the parent module
git_rev := $(shell git rev-parse --short HEAD)
git_tag := $(shell git tag --points-at=$(git_rev) | egrep -o "[0-9]+\.[0-9]+\.[0-9]+")
version := $(if $(git_tag),v$(git_tag),dev-$(git_rev))
release_version=$(if $(releaseVersion),$(releaseVersion),$(version))
build_time := $(shell date -u)
ldflags := -X "github.com/supreethrao/support-bot/cmd.version=$(release_version)" -X "github.com/supreethrao/support-bot/cmd.buildTime=$(build_time)"

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

.PHONY: setup check check-os vet lint format test build build-all-platforms install docker-local clean

setup:
	@echo "== setup"
	go get -v golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports

check : check-os vet lint format
	@[ "${moduleDir}" ] || ( echo "moduleDir needs to be passed as argument for this target. make moduleDir=<dirName> TARGET"; exit 1 )

check-os:
ifndef target_os
	$(error Unsupported platform: ${os})
endif

vet :
	@echo "== vet"
	@go vet $(pkgs)

lint :
	@echo "== lint"
	@for pkg in $(pkgs); do \
		golint -set_exit_status $$pkg || exit 1 ; \
	done;

format :
	@echo "== format"
	@goimports -w $(files)
	@sync

test :
	@echo "== run tests"
	go test -v -race $(pkgs)

build : check test
	@echo "== build"
	GOOS=${target_os} GOARCH=amd64 go build -ldflags '-s $(ldflags)' -o ${build_dir}/${target_os}/next-to-support -v

build-all-platforms : check test
	@echo "== building binary for linux amd64"
	GOOS=${target_os} GOARCH=amd64 go build -ldflags '-s $(ldflags)' -o ${build_dir}/${target_os}/next-to-support -v
	@echo "Cross compiling and building binary for ${cross_os} amd64"
	GOOS=${cross_os} GOARCH=amd64 go build -ldflags '-s $(ldflags)' -o ${build_dir}/${cross_os}/next-to-support -v
	@echo "Cross compiling binary for windows amd64"
	GOOS=windows GOARCH=amd64 go build -ldflags '-s $(ldflags)' -o ${build_dir}/windows/next-to-support.exe -v

install : build-all-platforms
	@echo "== installing binaries to dist folder"
	@mkdir -p $(dist_dir)
	@for platform in $(PLATFORMS); \
	do \
		echo "Installing $${platform} binary"; \
		cp -r $(build_dir)/$${platform} $(dist_dir)/$${platform}; \
	done;

docker-local: build
	@echo "docker"
	@echo "$(version)"
	@cd ${moduleDir}
	@docker build -t local/support-bot:v$(version) .

clean :
	@echo "== clean"
	rm -rf build
