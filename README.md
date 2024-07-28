# `ssoctx`: AWS SSO context switcher

[![Go Test Status](https://github.com/catpaladin/ssoctx/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/catapaladin/ssoctx/actions/workflows/test.yml)

Easily set and change AWS SSO context.

# Installation

## linux / macOS

```
curl -L https://raw.githubusercontent.com/catpaladin/ssoctx/main/scripts/install.sh | bash
```

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
  refresh     Refresh your previously used credentials
  select      Login to AWS SSO and select account and role
  version     Print the version number of the application

Flags:
  -h, --help   help for ssoctx

Use "ssoctx [command] --help" for more information about a command.

```

## `generate config`
```
ssoctx config generate
```

- Enter the SSO URL for your Identity Center URL
- Enter the region for that the URL is setup on

## `select`
```
ssoctx select
```

This will open a browser and login to AWS SSO, if your cached access is expired.
It will default to the custom `credential_process`.

This may not be ideal, if you're looking to do something like mount your credentials to a docker volume where the binary does not exist.
Use the `--persist` flag to write out the temporary keys.

Use the `--profile` flag to set credentials to a profile.

```
Login to AWS SSO by retrieving short-lived credentials for account and role.

Usage:
  ssoctx select [flags]

Flags:
  -a, --account-id string   set account id for desired aws account
      --clean               toggle if you want to remove lock and access token
      --debug               toggle if you want to enable debug logs
  -h, --help                help for select
      --json                toggle if you want to enable json log output
      --keys                toggle if you want to write access/secret keys to credentials file
      --print-creds         outputs the credentials to stdout and not modifying credentials file
  -p, --profile string      the profile name to set in credentials file (default "default")
  -r, --region string       set / override aws region
  -n, --role-name string    set with permission set role name
  -u, --start-url string    set / override aws sso url start url
```

## `refresh`
```
ssoctx refresh
```

This will refresh the process to write out the credentials, refreshing access.

```
Refreshes the previously used credentials to the default profile.
  Use the flags to refresh profile credentials.

Usage:
  ssoctx refresh [flags]

Flags:
  -a, --account-id string   set account id for desired aws account
      --debug               toggle if you want to enable debug logs
  -h, --help                help for refresh
      --json                toggle if you want to enable json log output
      --keys                toggle if you want to write access/secret keys to credentials file
  -p, --profile string      the profile name to set in credentials file (default "default")
  -n, --role-name string    set with permission set role name
```
