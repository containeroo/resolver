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

- **`env:`** - Environment variables.
  Example:

  ```text
  env:PATH
  ```

  → value of `$PATH`.

- **`file:`** - Simple key-value files. Supports `KEY=VAL` lines, with optional `export` prefixes and `#` comments.
  Example:

  ```text
  file:/config/app.txt//USERNAME
  ```

  → value of `USERNAME` in `app.txt`.

- **`json:`** - JSON files. Supports dot-notation for nested keys and array indexing.
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

- **`yaml:`** - YAML files. Same dot/array/filter notation as JSON.
  Example:

  ```text
  yaml:/config/app.yaml//servers.[host=example.org].port
  ```

  → `"port"` of the object where `"host" == "example.org"`.

- **`ini:`** - INI files. Supports section+key or default section.
  Example:

  ```text
  ini:/config/app.ini//Database.User
  ```

  → value of `User` in `[Database]`.

  ```text
  ini:/config/app.ini//Key1
  ```

  → value of `Key1` in the `[DEFAULT]` section.

- **`toml:`** - TOML files. Dot-notation for nested keys and array indexing.
  Example:

  ```text
  toml:/config/app.toml//server.host
  ```

  → `"host"` under `[server]`.

- **No prefix** - Returns the value unchanged.
  Example:

  ```text
  just-a-literal
  ```

  → `"just-a-literal"`.

Here's a drop-in README section you can paste under **Usage** (or create a new "Batch resolution" section). It documents both helpers in the same tone/style as the rest of your README.

## Batch resolution

When you need to resolve a list of strings (e.g., CLI args, YAML arrays), use the slice helpers. Both preserve order, return a **new** slice, and leave inputs unchanged. Unknown schemes still **pass through** unchanged, just like `ResolveVariable`.

### `ResolveSlice` (strict)

```go
func ResolveSlice(values []string) ([]string, error)
```

Resolves each element using the default registry. If any element fails, the function **stops at the first error** and returns it. No partial results are returned.

- Stable order, input unchanged.
- Unknown schemes are returned as-is.
- Empty input returns an empty slice.

**Example:**

```go
vals := []string{
  "env:USER",                  // resolved from env
  "just-a-literal",            // unchanged
  "json:/cfg/app.json//host",  // resolved from file
}

out, err := resolver.ResolveSlice(vals)
if err != nil {
  // handle error
}
fmt.Println(out)
```

> Prefer `(*Registry).ResolveSlice` if you use a custom registry:
> `r.ResolveSlice(vals)`

### `ResolveSliceBestEffort`

```go
func ResolveSliceBestEffort(values []string) ([]string, []error)
```

Attempts to resolve **all** elements and never fails fast. Returns:

- `out`: a slice with the same length as the input (resolved values where possible).
- `errs`: one error **per failed element**, in input order. Error messages include the failing **index** and original token for easy debugging (e.g., `index 2 ("json:/cfg/..."): <reason>`).

For failed indices, `out[i]` is set to the zero value `""` (so you can use `errs` to decide how to fill defaults or report problems).

**Example:**

```go
vals := []string{
  "env:USER",                     // ok
  "secret:API_KEY",               // suppose this scheme errors
  "unknown:raw",                  // unknown → pass-through (no error)
}

out, errs := resolver.ResolveSliceBestEffort(vals)
// out[0] == resolved USER
// out[1] == "" (failed)
// out[2] == "unknown:raw" (unchanged)
for _, e := range errs {
  fmt.Println("resolve error:", e)
}
```

> As with the strict variant, there's also a registry method:
> `r.ResolveSliceBestEffort(vals)`

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
