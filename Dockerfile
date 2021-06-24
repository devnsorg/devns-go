FROM golang:1.16-alpine as BUILD
RUN apk add make --no-cache
RUN mkdir /app
WORKDIR /app

ARG MODULE=mustset
ARG VERSION=0.0.0
COPY $MODULE/go.mod /app/$MODULE/
COPY $MODULE/go.sum /app/$MODULE/
# For go.mod `replace`
COPY ./util /app/util/


RUN ls /app
WORKDIR /app/$MODULE/
RUN go mod download

ADD ./ /app
RUN make build_linux

RUN ls /app/$MODULE/bin/*

FROM alpine
RUN mkdir /app
WORKDIR /app
ARG MODULE=mustset
ARG VERSION=0.0.0
COPY --from=BUILD /app/$MODULE/bin/ /app/
RUN ls
RUN chmod a+x *

ENV DEVNS_MODULE $MODULE
ENV DEVNS_VERSION $VERSION
ENTRYPOINT /app/$DEVNS_MODULE-amd64-linux-$DEVNS_VERSION