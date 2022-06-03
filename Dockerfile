FROM docker.io/library/golang:1.17-alpine3.16 AS builder

LABEL maintainer="supreeth.gururaj@uttara.co.uk"

RUN apk add make git openssh

ENV GOPRIVATE "github.com/supreethrao/automated-rota-manager"
RUN git config --global url.git@github.com:.insteadOf https://github.com/
RUN mkdir -p -m 0700 /root/.ssh && ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

WORKDIR /repo/automated-rota-manager
COPY . .

RUN GOPATH="" make build

FROM docker.io/library/alpine:3.16

COPY --from=builder /repo/automated-rota-manager/build/linux/automated-rota-manager /app/automated-rota-manager

ENTRYPOINT /app/automated-rota-manager
