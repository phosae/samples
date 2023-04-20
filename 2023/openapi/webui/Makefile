publish:
	docker buildx build --platform linux/amd64,linux/arm64 -t zengxu/openapi-ui --push .

dev:
	docker buildx build -t zengxu/openapi-ui --load . && docker run --rm -p 8000:8000 zengxu/openapi-ui