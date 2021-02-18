FROM golang:1.16-buster AS builder

WORKDIR /go/src/app
COPY . /go/src/app

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM scratch
COPY --from=builder /go/bin/app /
CMD ["/app"]