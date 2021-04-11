FROM golang:1.16 as builder

WORKDIR /go/src/bind-host
RUN git clone --depth 1 https://github.com/warm-metal/bind-host.git .
RUN go mod download
RUN CGO_ENABLED=0 go build -o bind-host ./cmd/bind

FROM scratch
COPY --from=builder /go/src/bind-host/bind-host /
ENTRYPOINT ["/bind-host"]