FROM golang:1.16 as builder

WORKDIR /go/src/bind-host
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 go build -o bind-host ./cmd/bind

FROM scratch
COPY --from=builder /go/src/bind-host/bind-host /
ENV HOST_ROOTFS=""
ENV CRI_ADDR=""
ENV FSTAB=""
ENTRYPOINT ["/bind-host", "-rootfs=${HOST_ROOTFS}", "-cri-image=${CRI_ADDR}", "-fstab=${FSTAB}", "--"]