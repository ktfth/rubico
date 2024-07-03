# Rubico

## Description

The main goal of this project is to create an app focused on "magic" authentication, serving as a base for microservices.

## Installation

Rubico depends on `serviceweaver`, check out the documentation to install the cli.

https://serviceweaver.dev/

## Running the project

```
go mod tidy
weaver generate .
SERVICEWEAVER_CONFIG=weaver.toml go run .
```

## Considerations

This is a small app, who use resend as a service for authentication with a magic link.
