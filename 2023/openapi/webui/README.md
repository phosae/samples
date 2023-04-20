# Local OpenAPI Viewer
![img](./openapi_viewer.png)
## play
run in docker and visit http://localhost:8000 in browser
```shell
docker run -it --rm -p 8000:8000 --entrypoint bash zengxu/openapi-ui
```
## build
```shell
docker buildx build --platform linux/amd64,linux/arm64 -t zengxu/openapi-ui --push .
```
