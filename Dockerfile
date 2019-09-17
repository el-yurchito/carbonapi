FROM registry.k.avito.ru/avito/golang:1.11 as builder

WORKDIR /go/src/github.com/go-graphite/carbonapi
COPY . .

RUN apt-get update
RUN apt-get install -y make gcc git
RUN make test-nocairo
RUN make nocairo


FROM registry.k.avito.ru/avito/debian-minbase:latest

WORKDIR /
COPY --from=builder /go/src/github.com/go-graphite/carbonapi/carbonapi ./usr/bin/carbonapi

RUN apt-get install -y ca-certificates

CMD ["carbonapi", "-config", "/etc/carbonapi.yml"]
