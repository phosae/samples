FROM --platform=$BUILDPLATFORM golang:1.19 as builder
ARG TARGETOS TARGETARCH
WORKDIR /workspace

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY main.go main.go
COPY middleware.go middleware.go

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags '-s -w -extldflags "-static"' -trimpath -a -o app-$TARGETARCH .

FROM ubuntu:jammy
ARG TARGETARCH
WORKDIR /go-http-range
COPY --from=builder /workspace/app-$TARGETARCH /go-http-range/app 
COPY /media /go-http-range/media

ENTRYPOINT ["/go-http-range/app"]
