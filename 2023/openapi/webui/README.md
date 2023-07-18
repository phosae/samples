# Local OpenAPI Viewer
![img](./openapi_viewer.png)

## play

run demo in docker and visit http://localhost:8000 in browser

```shell
docker run --rm -p 8000:8000 zengxu/openapi-ui
```

Load custom OpenAPI JSON/YAML files by

```shell
docker run --rm -v /openapi/spec/dir/:/static/spec/ -p 8000:8000 zengxu/openapi-ui

# or 

docker run --rm -v /openapi/spec/dir/swagger.json:/static/spec/swagger.json -p 8000:8000 zengxu/openapi-ui
```

The Local OpenAPI Viewer can also load a remote OpenAPI JSON/YAML file by adding a URL argument

```shell
docker run --rm -p 8000:8000 zengxu/openapi-ui https://raw.githubusercontent.com/phosae/x-kubernetes/a0e515df668c02b16e8123fbb79b06e2c6a09b0a/apiserver-from-scratch/docs/swagger.json
```

## dev
```shell
make dev
```

## publish
```shell
make
```
