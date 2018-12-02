all: check install docker

imageVersion=PASSED_FROM_CLI
currentGoPath=~/workspace/blah

setup:
	@echo "setup"
	@go get github.com/golang/lint/golint
	@go get golang.org/x/tools/cmd/goimports
	@go get github.com/golang/dep/cmd/dep
	dep ensure

check:
	@echo "check"
	@ginkgo rota_test

install:
	@echo "install"
	@mkdir -p ./bin
	@env GOOS=linux GOARCH=amd64 go install -v ./next-to-support
	@cp $(currentGoPath)/bin/next-to-support ./bin
	

docker:
	@echo "docker"
	@echo "$(imageVersion)"
	@docker build -t local/support-bot:v$(imageVersion) .
	@docker tag local/support-bot:v$(imageVersion) registry.tools.cosmic.sky/core-engineering/test-repo/support-bot:v$(imageVersion)
	@docker push registry.tools.cosmic.sky/core-engineering/test-repo/support-bot:v$(imageVersion)
