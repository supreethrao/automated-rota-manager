FROM phusion/baseimage

LABEL maintainer="supreeth.gururaj@uttara.co.uk"

COPY bin/next-to-support /app/next-to-support

ENTRYPOINT /app/next-to-support
