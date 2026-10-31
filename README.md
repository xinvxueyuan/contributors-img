# contributors-img

A contributor card rendered as an SVG, served by a single Go service on Vercel.
Embed it in any README:

```html
<img src="https://contributors.hikarilink.org/image?repo=owner/name" />
```

`https://contributors.hikarilink.org/image?repo=owner/name`

## Parameters

| Name | Default | Description |
| --- | --- | --- |
| `repo` | — (required) | `owner/name` of a public GitHub repository. |
| `max` | `100` | Maximum number of contributors to draw. |
| `columns` | `12` | Number of avatars per row. |
| `anon` | `false` | `true` includes contributors without a GitHub account. |

Responses carry `cache-control: public, max-age=259200` (3 days) and an `etag`,
and answer `If-None-Match` with `304`, so an embedded image costs at most one
render every three days.

## Running locally

```bash
GITHUB_AUTH_TOKEN=<a GitHub token> go run .
curl -sS "http://localhost:3333/image?repo=angular/angular-ja&max=5"
```

## Deploying

The repository is a single Go service: `main.go` listens on `PORT`, so Vercel's
Go framework preset builds and runs it as-is (`vercel.json` pins the preset and
the region). Two environment variables are required in the Vercel project:

| Name | Value | Notes |
| --- | --- | --- |
| `GITHUB_AUTH_TOKEN` | a GitHub token | Read-only public scope is enough; 5000 requests/hour. |
| `APP_ENV` | `production` | Turns on JSON logs and gin release mode. |

Alternatively the GitHub App triple (`GITHUB_APP_ID`, `GITHUB_APP_INSTALLATION_ID`
and `GITHUB_APP_PRIVATE_KEY`) can be set instead of `GITHUB_AUTH_TOKEN`.

This project was bootstrapped from a deployment, so Vercel does not redeploy on
push yet: connect the repository once in Project → Settings → Git, after which
every push to `main` ships.

Rendered images and contributor lists are kept in an in-process LRU cache (300
entries), so repeated requests are free and GitHub API usage stays low without
an external store.

## What this fork drops

Upstream (`lacolaco/contributors-img`) also ships an Angular preview site, a
Firestore-backed gallery, BigQuery usage logging, Cloud Trace and a Cloud Run
deployment. None of that is needed to serve a card, so this fork keeps only the
image endpoint and runs entirely on Vercel.

