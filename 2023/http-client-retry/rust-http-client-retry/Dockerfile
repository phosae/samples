FROM rust:1.74.0-slim-bullseye AS build
WORKDIR /workspace
ARG TARGETARCH

# start of mirror settings for Chinese users
ENV RUSTUP_DIST_SERVER https://mirrors.ustc.edu.cn/rust-static 
ENV RUSTUP_UPDATE_ROOT https://mirrors.ustc.edu.cn/rust-static/rustup
RUN <<EOF bash
printf '[source.crates-io]
replace-with = "ustc"
[source.ustc]
registry = "sparse+https://mirrors.ustc.edu.cn/crates.io-index/"
' | tee -a ${CARGO_HOME:-$HOME/.cargo}/config
EOF
# end of mirror settings for Chinese users

COPY platform.sh platform.sh
COPY Cargo.toml Cargo.toml
COPY src/ src/

RUN ./platform.sh 
RUN rustup target add $(cat /.platform) 
RUN rustup toolchain install stable-$(cat /.platform) 

RUN apt update && apt install -y pkg-config openssl libssl-dev

RUN --mount=type=cache,target=/usr/local/cargo/git/db \
    --mount=type=cache,target=/usr/local/cargo/registry/cache \
    --mount=type=cache,target=/usr/local/cargo/registry/index \
    RUST_BACKTRACE=1 cargo build --target $(cat /.platform) --release

RUN cp /workspace/target/$(cat /.platform)/release/rcli /usr/local/bin/rcli

FROM debian:bullseye-slim
COPY --from=build /usr/local/bin/rcli /rcli
ENTRYPOINT ["/rcli"]