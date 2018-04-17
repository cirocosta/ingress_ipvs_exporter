FROM golang:alpine as builder

ADD ./ /go/src/github.com/cirocosta/ipvs_exporter
WORKDIR /go/src/github.com/cirocosta/ipvs_exporter

RUN set -ex && \
  CGO_ENABLED=0 go build -tags netgo -v -a -ldflags '-extldflags "-static"' && \
  mv ./ipvs_exporter /usr/bin/ipvs_exporter

FROM alpine
COPY --from=builder /usr/bin/ipvs_exporter /usr/local/bin/ipvs_exporter

ENTRYPOINT [ "ipvs_exporter" ]

