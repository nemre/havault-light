# havault-light
Lightest HashiCorp Vault Go Client Library

### Usage
```go
package main

import (
    "context"

    "github.com/nemre/havault-light"
)

func main() {
    ctx := context.TODO()

    vault, err := havaultlight.New(ctx, &havaultlight.Config{
        Addr: "http://localhost:8200",
        Engine: "secret",
        Token: "foo",
    })
    if err != nil {
        panic(err)
    }

    if err = vault.Ping(ctx); err != nil {
        panic(err)
    }

    if err = vault.Set(ctx, "foo", map[string]any{"a": "b"}); err != nil {
        panic(err)
    }

    foo, err := vault.Get(ctx, "foo")
    if err != nil {
        panic(err)
    }

    println(foo["a"])

    if err = vault.Delete(ctx, "foo"); err != nil {
        panic(err)
    }
}
