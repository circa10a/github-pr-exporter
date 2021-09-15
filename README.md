# <img src="https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png" height="5%" width="5%" align="left"/> github-pr-exporter

A prometheus exporter for monitoring pull requests for specified users in the last X number of days. Useful for tracking things like [hacktoberfest](https://hacktoberfest.digitalocean.com/) within your org.

![Build Status](https://github.com/circa10a/github-pr-exporter/workflows/deploy/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/circa10a/github-pr-exporter)](https://goreportcard.com/report/github.com/circa10a/github-pr-exporter)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/circa10a/github-pr-exporter?style=plastic)
![Docker Pulls](https://img.shields.io/docker/pulls/circa10a/github-pr-exporter?style=plastic)

![alt text](https://i.imgur.com/Ur1N1S5.png)

## Usage

Once started, application will be accessible at http://localhost:8080/metrics

### CLI

```log
# Install with Go
go install github.com/circa10a/github-pr-exporter@latest

# Execute
❯ github-pr-exporter -h
Usage of ./github-pr-exporter:
      --config string       Path to config file (default "./config.yaml")
      --days-ago int        How many days back to search for pull requests (default 90)
      --ignore-user-repos   Ignore the user's own repos
      --interval int        How many seconds to wait before refreshing pull request data. Defaults to 6 hours (default 21600)
      --port int            What port to listen on (default 8080)
pflag: help requested
```

#### Configuration file

The CLI expects a YAML config file like so:

```yaml
---
config:
  users:
    - circa10a
    - sindresorhus
```

You can also look in the [examples](/examples) directory

### Docker

```bash
docker run -p 8080:8080 -v $PWD/config.yaml:/config.yaml circa10a/github-pr-exporter
```

#### docker-compose

First, update `examples/config.yaml`

Then, to start a preconfigured prometheus + grafana + exporter stack:

```bash
docker-compose up
```

Then you can browse the preconfigured dashboard at http://localhost:3000/d/h_PRluMnk/pull-requests?orgId=1

## Metrics

| Name               | Type  | Cardinality  | Help                                                                                                |
|--------------------|-------|--------------|-----------------------------------------------------------------------------------------------------|
| pull_request       | gauge | 4            | A pull request from a user in config. Contains labels `user`, `created_at`, `link`, and `status`    |
| pull_request_total | gauge | 0            | Total number of pull requests created by all users in the configured time window (previous 90 days) |
