.PHONY: media
media:
	wget https://github.com/phosae/bin/releases/download/range-mp4/tomato-egg_stir-fry.mp4 -P ./media

run:
	go run main.go middleware.go

build:
	CGO_ENABLED=0 go build -o go-http-range .

docker-build:
	docker buildx build -t zengxu/go-http-range --platform linux/amd64,linux/arm64 --push .
