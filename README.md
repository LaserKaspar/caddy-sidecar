# caddy-sidecar

Update and reload a Caddyfile remotely over HTTP. The tool provides an alternative API for manaing Caddyfile configuration instead of JSON. This helps preserve comments, imports, and the original Caddyfile structure. It is intended for remote administration, automation and deployment pipelines.

## motivation

Bridges the gap between Caddy’s two configuration paths:

- Caddyfile on disk - Supports comments, imports, and organization. However, applying changes requires shell and disk access.

- Admin API (JSON) - Accessible over HTTP, but JSON does not support comments or imports and breaks original Caddyfile configs.

caddy-sidecar combines the advantages of both approaches by exposing them over HTTP:

- A Caddyfile API for managing files under a configuration root.

- A POST /reload endpoint that reloads the running Caddy process.

Write the file, call /reload, done.

## setup

see [.env.exampe](.env.example) and [docker-compose.yml](docker-compose.yml)

## usage

Variable `<path>` can be replaced with `Caddyfile` for example

Get Caddyfile
```
curl -H "X-API-Key: api-secret-key" \
  http://localhost:8080/file/<path>
```

Create a new config file
```
curl -X PUT \
  -H "X-API-Key: api-secret-key" \
  --data-binary @Caddyfile.new \
  http://localhost:8080/file/<path>
```

Update/replace a config file
```
curl -X POST \
  -H "X-API-Key: api-secret-key" \
  --data-binary @Caddyfile.updated \
  http://localhost:8080/file/<path>
```

Delete a Caddyfile
```
curl -X DELETE -H "X-API-Key: api-secret-key" http://localhost:8080/file/Caddyfile
```

Reload the files on disk
```
curl -X POST -H "X-API-Key: api-secret-key" http://localhost:8080/reload
```
