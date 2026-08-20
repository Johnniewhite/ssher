# Project website

The source for [getssher.com](https://getssher.com) lives in `website/`. It is a
dependency-free static site. `deploy/website/` includes the production Nginx
virtual host and a standalone Docker Compose + Caddy alternative.

The ssher Cloud launch film is embedded from `website/ssher-cloud-launch.mp4`.
Its editable Remotion source and deterministic audio generator live in `video/`.

## Local preview

From the repository root:

```sh
python3 -m http.server 8080 --directory website
```

Open `http://localhost:8080`.

## Production layout

Production uses immutable release directories and a movable `current` symlink:

```text
/var/www/ssher-marketing/
├── current -> releases/<timestamp>
└── releases/
    └── <timestamp>/
```

Nginx terminates the Cloudflare origin connection, redirects `www` to the apex,
serves byte ranges for the launch film, and adds the security headers declared
in `deploy/website/nginx.conf`. Cloudflare proxies the apex and `www` records in
Full (strict) mode.

## Deploy

Copy `website/` into a new timestamped directory, atomically point `current` at
it, install `deploy/website/nginx.conf`, then validate and reload Nginx:

```sh
nginx -t
systemctl reload nginx
```

The certificate and key paths in the Nginx file expect a Cloudflare Origin CA
certificate covering `getssher.com` and `www.getssher.com`. Keep the private key
mode `0600`; never commit it. For a standalone host, the included Caddy setup can
obtain and manage a publicly trusted certificate instead.

## Verification

```sh
curl -I https://getssher.com
curl -I https://www.getssher.com
```

Check that the apex responds over HTTP/2 or HTTP/3, `www` redirects to the apex,
and the security headers are present.
