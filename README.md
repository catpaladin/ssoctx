# aws-sso-util

# Installation

## brew
* Create a Github [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).
  * Check the box for `repo` access.
  * Save your token somewhere safe.


```
# export the credentials
export HOMEBREW_GITHUB_API_TOKEN=<personal_access_token>

# brew tap the repo
brew tap catpaladin/aws-sso-util https://github.com/catpaladin/aws-sso-util

# brew install
brew install catpaladin/aws-sso-util/aws-sso-util
```

## go
requires go v1.18+

```
go install github.com/catpaladin/aws-sso-util@main
```

# Usage
```
A tool for seting up AWS SSO.
		Use to login to SSO portal and refresh session.

Usage:
  aws-sso-util [command]

Available Commands:
  assume      Assume directly into an account and SSO role
  completion  Generate the autocompletion script for the specified shell
  config      Handles configuration
  help        Help about any command
  login       Login to AWS SSO
  refresh     Refresh your previously used credentials

Flags:
  -h, --help   help for aws-sso-util

Use "aws-sso-util [command] --help" for more information about a command.
```

## generate config
```
aws-sso-util config generate
```

## login
```
aws-sso-util login
```
