.PHONY: media
media:
	wget https://github.com/phosae/bin/releases/download/range-mp4/dun-dun-dance-part1.mp4 -P ./media
	wget https://github.com/phosae/bin/releases/download/range-mp4/dun-dun-dance.mp4 -P ./media

run:
	go run main.go middleware.go

run-dynamic:
	go run main.go middleware.go -dynamic true

build:
	CGO_ENABLED=0 go build -o go-http-range .

docker-build:
	docker buildx build -t zengxu/go-http-range:dynamic --platform linux/amd64,linux/arm64 --push .
