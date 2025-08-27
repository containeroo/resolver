# Resolver Package

The `resolver` package provides a flexible and extensible way to resolve configuration values from various sources including environment variables, plain files, `JSON`, `YAML`, `INI`, `TOML`, and key-value files.
It uses a **prefix-based system** to determine which resolver to use, and returns either the resolved value or an error.

## Installation

```bash
go get github.com/containeroo/resolver/resolver
```

## Usage

The primary entry point is the `ResolveVariable` function.
It takes a string and attempts to resolve it based on its prefix:

- **`env:`** – Environment variables.
  Example:

  ```text
  env:PATH
  ```

  → value of `$PATH`.

- **`file:`** – Simple key-value files. Supports `KEY=VAL` lines, with optional `export` prefixes and `#` comments.
  Example:

  ```text
  file:/config/app.txt//USERNAME
  ```

  → value of `USERNAME` in `app.txt`.

- **`json:`** – JSON files. Supports dot-notation for nested keys and array indexing.
  Example:

  ```text
  json:/config/app.json//server.host
  ```

  → `"host"` under `"server"`.

  ```text
  json:/config/app.json//servers.0.host
  ```

  → `"host"` of the first element in `"servers"`.

  ```text
  json:/config/app.json//servers.[name=api].port
  ```

  → `"port"` of the object in `"servers"` where `"name" == "api"`.

- **`yaml:`** – YAML files. Same dot/array/filter notation as JSON.
  Example:

  ```text
  yaml:/config/app.yaml//servers.[host=example.org].port
  ```

  → `"port"` of the object where `"host" == "example.org"`.

- **`ini:`** – INI files. Supports section+key or default section.
  Example:

  ```text
  ini:/config/app.ini//Database.User
  ```

  → value of `User` in `[Database]`.

  ```text
  ini:/config/app.ini//Key1
  ```

  → value of `Key1` in the `[DEFAULT]` section.

- **`toml:`** – TOML files. Dot-notation for nested keys and array indexing.
  Example:

  ```text
  toml:/config/app.toml//server.host
  ```

  → `"host"` under `[server]`.

- **No prefix** – Returns the value unchanged.
  Example:

  ```text
  just-a-literal
  ```

  → `"just-a-literal"`.

## Example

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/containeroo/resolver"
)

func main() {
    os.Setenv("MY_VAR", "HelloWorld")

    val, err := resolver.ResolveVariable("env:MY_VAR")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(val) // Output: HelloWorld

    // Example: resolve a JSON key
    // File: /config/app.json
    // {
    //   "server": {
    //     "host": "localhost",
    //     "port": 8080
    //   }
    // }
    host, err := resolver.ResolveVariable("json:/config/app.json//server.host")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(host) // Output: localhost
}
```

## Extensibility

You can register your own resolver schemes at runtime:

```go
resolver.RegisterResolver("secret:", myCustomResolver)
```

Resolvers must implement:

```go
type Resolver interface {
    Resolve(value string) (string, error)
}
```

This allows you to plug in custom backends (e.g., Vault, Consul, HTTP endpoints).
