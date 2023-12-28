## Bad Server

`docker run -p 8080:8080 --rm zengxu/bad-server` run server in standalone mode 
- `/`, abort connection without reply
- `/429`, response with `429 TooManyRequests`
- `/500`, response with `500 InternalServerError`
- `/503`, response with `503 ServiceUnavailable`

`docker run -e MODE=none --rm zengxu/bad-server` run none server

`docker run -e MODE=proxy -p 8080:8080 --rm zengxu/bad-server` run server with reverse proxy behind
- `/`, response with `502 BadGateway`
- `/none`, response with `502 BadGateway`
- `/429`, response with `429 TooManyRequests`
- `/500`, response with `500 InternalServerError`
- `/503`, response with `503 ServiceUnavailable`
- `/504`, response with `504 GatewayTimeout`
