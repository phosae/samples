FROM --platform=$BUILDPLATFORM golang:1.21 as builder
ARG TARGETOS TARGETARCH
WORKDIR /workspace

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.mod
RUN go mod download
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -a -o bad-server-$TARGETARCH .

FROM ubuntu:jammy
ARG TARGETARCH
WORKDIR /
COPY --from=builder /workspace/bad-server-$TARGETARCH /bad-server

ENTRYPOINT ["/bad-server"]