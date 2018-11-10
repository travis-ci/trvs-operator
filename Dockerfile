FROM golang:1.11 AS builder

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

RUN update-ca-certificates

WORKDIR $GOPATH/src/github.com/travis-ci/trvs-operator
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /trvs-operator .

FROM ruby:2.5-alpine
COPY --from=builder /trvs-operator .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./trvs-operator"]
