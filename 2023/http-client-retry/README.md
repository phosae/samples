# HTTP client retry examples

start [bad server](./bad-server) for testing

```bash
docker run -e MODE=proxy -dp 8080:8080 --rm zengxu/bad-server
```

## Rust client


If Rust env have been ready on your machine, just

```bash
cd rust-http-client-retry/ && cargo run http://localhost:8080
```

or here're docker image for playing

```bash
docker run --network host --rm zengxu/rust-http-client-retry http://localhost:8080
```

## Golang Client

```bash
cd golang-http-client-retry/ && go run main.go http://localhost:8080
```