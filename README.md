# `ssoctx`: AWS SSO context switcher

[![Go Test Status](https://github.com/catpaladin/ssoctx/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/catapaladin/ssoctx/actions/workflows/test.yml)

Easily set and change AWS SSO context.

# Installation

## linux / macOS

## go
requires go v1.22+

```
go install github.com/catpaladin/ssoctx/cmd/ssoctx@latest
```

# Usage
```
A tool for seting up AWS SSO.
Use to login to SSO portal and refresh session.

Usage:
  ssoctx [command]

Available Commands:
  assume      Assume directly into an account and SSO role
  completion  Generate the autocompletion script for the specified shell
  config      Handles configuration
  help        Help about any command
  login       Login to AWS SSO
  refresh     Refresh your previously used credentials
  version     Print the version number of the application

Flags:
  -h, --help   help for ssoctx

Use "ssoctx [command] --help" for more information about a command.

```

## generate config
```
ssoctx config generate
```

## login
```
ssoctx login
```
