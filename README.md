# psctl

Polar Signals CLI.

## Installation

```bash
go install github.com/polarsignals/psctl@latest
```

## Usage

### Login

Authenticate with Polar Signals using OAuth device flow:

```bash
psctl auth login
```

Use `--no-open` to print the URL instead of opening a browser automatically:

```bash
psctl auth login --no-open
```

### Check authentication status

```bash
psctl auth status
```

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
