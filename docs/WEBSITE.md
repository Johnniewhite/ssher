# Project website

The source for [getssher.com](https://getssher.com) lives in `website/`. It is a
dependency-free static site served by Caddy from the deployment files in
`deploy/website/`.

## Local preview

From the repository root:

```sh
python3 -m http.server 8080 --directory website
```

Open `http://localhost:8080`.

## Production layout

The server keeps a deployment checkout under `/opt/ssher-web`:

```text
/opt/ssher-web/
├── Caddyfile
├── compose.yml
└── site/
```

Caddy terminates HTTPS, redirects `www` to the apex domain, compresses static
assets, and adds the security headers declared in `Caddyfile`. Its certificate
and configuration state are stored in named Docker volumes.

## Deploy

Copy `website/` to `/opt/ssher-web/site/`, copy the two files in
`deploy/website/` to `/opt/ssher-web/`, then validate and restart:

```sh
docker compose -f /opt/ssher-web/compose.yml config --quiet
docker compose -f /opt/ssher-web/compose.yml up -d
docker compose -f /opt/ssher-web/compose.yml ps
```

Before requesting a certificate, the DNS records for `getssher.com` and
`www.getssher.com` must resolve to the production server. When Cloudflare proxying
is enabled, its SSL/TLS mode should be **Full (strict)** after Caddy obtains the
origin certificate.

## Verification

```sh
curl -I https://getssher.com
curl -I https://www.getssher.com
```

Check that the apex responds over HTTP/2 or HTTP/3, `www` redirects to the apex,
and the security headers are present.
