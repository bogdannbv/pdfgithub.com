# [pdfgithub.com](https://pdfgithub.com)

Open GitHub PDF files directly in the browser by prefixing `github.com` with
`pdf`.

```diff
- https://github.com/bogdannbv/pdfgithub.com/blob/main/example/file.pdf
+ https://pdfgithub.com/bogdannbv/pdfgithub.com/blob/main/example/file.pdf
```

The app rewrites GitHub PDF URLs to GitHub's raw content host and
streams the PDF response back to the browser.

## Usage

1. Find a GitHub URL that points to a `.pdf` file.
2. Replace `github.com` with `pdfgithub.com`.
3. Open the new URL in your browser.

The root (`/`) page shows a small usage guide.

## Run Locally

Create a `.env` file:

```sh
cp .env.example .env
```

Run the app:

```sh
go run .
```

With the default example config, open:

```text
http://localhost:8080
```

## Configuration

| Variable         | Description                                           | Default            |
|------------------|-------------------------------------------------------|--------------------|
| `HTTP_BIND_HOST` | Host address the HTTP server binds to.                | `0.0.0.0`          |
| `HTTP_BIND_PORT` | Port the HTTP server listens on.                      | `80`               |
| `APP_ENV`        | Set to `local` to use debug-level logging.            | info-level         |
| `GH_TOKEN`       | GH Personal access token with public repo read access | none, **required** |
