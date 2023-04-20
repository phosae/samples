# web

dev, build in Docker

```
docker run -p 5173:5173 --rm -it -v $PWD:/web -w /web node:18.16.0-bullseye-slim bash
```

## Customize configuration

See [Vite Configuration Reference](https://vitejs.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev -- --host
```

### Compile and Minify for Production

```sh
npm run build
```
