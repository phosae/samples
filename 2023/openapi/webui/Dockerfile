FROM --platform=$BUILDPLATFORM  node:18.16.0-bullseye-slim as webbuilder

COPY /web /web
WORKDIR /web
RUN npm install
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.18 as gobuilder
ARG TARGETOS TARGETARCH

WORKDIR /workspace
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.mod
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o web-manager-$TARGETARCH main.go

FROM debian:bullseye
ARG TARGETARCH
WORKDIR /
COPY --from=webbuilder /web/dist /static
COPY --from=gobuilder /workspace/web-manager-$TARGETARCH /web-manager
COPY /spec /static/spec
USER 65532:65532
EXPOSE 8000

ENV SPEC_DIR=/static
ENTRYPOINT ["/web-manager"]
