FROM golang:alpine as build-img

RUN apk update && \
    apk add musl-dev ca-certificates make git

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY $PWD .
RUN make build

FROM alpine
COPY --from=build-img /build/promdiscovery /promdiscovery

ENTRYPOINT [ "/promdiscovery" ]