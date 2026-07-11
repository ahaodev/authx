# authx

authx is a lightweight Go library for integrating OAuth-based third-party login providers into your own backend service.

It is designed for projects that already have their own user system, session handling, and RBAC/permission logic. authx focuses only on:

- redirecting users to the provider
- handling the callback
- exchanging the authorization code for an access token
- fetching the user profile
- returning a unified profile model for your application

This library intentionally stays small and framework-agnostic.

## Why authx?

If you already have your own user table, JWT logic, and permission system, you usually do not want to introduce a heavy IAM system just for social login.

authx helps you:

- keep authentication logic separate from business logic
- support multiple providers through a common interface
- reuse the same flow for Google, GitHub, WeChat, Alipay, and more
- keep your service lightweight and easy to maintain

## Features

- provider-agnostic abstraction
- OAuth2-based authorization flow
- state generation and validation
- callback handling
- unified user profile model
- easy extension for new providers
- no dependency on a specific web framework
- no database dependency in the core library

## Installation

```bash
go get github.com/ahaodev/authx
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/ahaodev/authx"
    _ "github.com/ahaodev/authx/providers/google"
)

func main() {
    ctx := context.Background()

    provider, err := authx.NewProvider("google", map[string]string{
        "client_id":     "your-client-id",
        "client_secret": "your-client-secret",
        "redirect_url":  "http://localhost:8080/auth/callback",
    })
    if err != nil {
        panic(err)
    }

    flow := authx.NewFlow(provider, authx.NewMemoryStateStore())
    authURL, state, err := flow.Start()
    if err != nil {
        panic(err)
    }

    fmt.Println("Open this URL:", authURL)
    fmt.Println("State:", state)

    code := "received-code-from-provider"
    profile, err := flow.HandleCallback(ctx, code, state)
    if err != nil {
        panic(err)
    }

    fmt.Printf("User profile: %+v\n", profile)
}
```

## Core Concepts

### Provider

A provider is a specific integration implementation such as Google or GitHub.

### Profile

A unified profile object returned by the provider:

```go
type Profile struct {
    Provider  string
    Subject   string
    Email     string
    Name      string
    AvatarURL string
    Raw       map[string]interface{}
}
```

### Flow

The flow handles the common OAuth process:

- generate state
- build authorize URL
- handle callback
- exchange code for token
- fetch profile

## Architecture

authx is intentionally designed as a thin library:

- your service owns users, roles, permissions, sessions
- authx only handles identity-provider integration
- your application decides how to map the returned profile to your own user model

This keeps the library generic and reusable.

## How to Add a New Provider

To add a new provider, implement the `Provider` interface:

```go
type Provider interface {
    Name() string
    AuthURL(state string) (string, error)
    Exchange(ctx context.Context, code string) (*Token, error)
    UserInfo(ctx context.Context, token *Token) (*Profile, error)
}
```

Then register it in the provider factory.

## Roadmap

- support Google
- support GitHub
- support WeChat
- support Alipay
- support Microsoft
- support Apple
- add Redis-backed state store
- add examples for Gin

## License

MIT

## Contributing

Contributions are welcome. Please open an issue or submit a pull request with a clear description of the change.
