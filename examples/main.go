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
