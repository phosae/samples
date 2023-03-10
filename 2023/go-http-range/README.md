# How To
download media
```
make media
```

run
```
go run main.go middleware.go

or 

make run
```

build
```
CGO_ENABLED=0 go build -o go-http-range .

or 

make build
```
docker build
 
```
docker buildx build -t zengxu/go-http-range --platform linux/amd64,linux/arm64 --push .

or

make docker-build
```
