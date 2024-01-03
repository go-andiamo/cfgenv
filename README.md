# CFGENV
[![GoDoc](https://godoc.org/github.com/go-andiamo/cfgenv?status.svg)](https://pkg.go.dev/github.com/go-andiamo/cfgenv)
[![Latest Version](https://img.shields.io/github/v/tag/go-andiamo/cfgenv.svg?sort=semver&style=flat&label=version&color=blue)](https://github.com/go-andiamo/cfgenv/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-andiamo/cfgenv)](https://goreportcard.com/report/github.com/go-andiamo/cfgenv)

Cfgenv loads config structs from environment vars.

Struct field types supported:
* _native type_ - `string`, `bool`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `time.Duration`
* _pointer native type_ - `*string`, `*bool`, `*int`, `*int8`, `*int16`, `*int32`, `*int64`, `*uint`, `*uint8`, `*uint16`, `*uint32`, `*uint64`, `*float32`, `*float64`, `*time.Duration` - _environment var is optional and value is not set if the env var is missing_
* `[]V` _(slice)_ where `V` is _native type_ or _pointer native type_
* `map[K]V` where `K` is _native type_ and `V` is _native type_ or _pointer native type_
* embedded structs
* other types can be handled by providing a `cfgenv.CustomerSetterOption`

Example:

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/cfgenv"
)

type DbConfig struct {
    Host     string `env:"optional,default=localhost"`
    Port     uint   `env:"optional,default=3601"`
    Username string
    Password string
}

type Config struct {
    ServiceName string
    Database    DbConfig `env:"prefix=DB"`
}

func main() {
    cfg := &Config{}
    err := cfgenv.Load(cfg)
    if err != nil {
        panic(err)
    } else {
        fmt.Printf("%+v\n", cfg)
    }
}
```
would effectively load from environment...
```
SERVICE_NAME=foo
DB_HOST=localhost
DB_PORT=33601
DB_USERNAME=root
DB_PASSWORD=root
```

## Installation
To install cfgenv, use go get:

    go get github.com/go-andiamo/cfgenv

To update cfgenv to the latest version, run:

    go get -u github.com/go-andiamo/cfgenv

## Tags

Fields in config structs can use the `env` tag to override cfgenv loading behaviour

| Tag                 | Purpose                                                                                                           |
|---------------------|-------------------------------------------------------------------------------------------------------------------|
| `env:"MY"`          | overrides the environment var name to read with `MY`                                                              |
| `env:"optional"`    | denotes the environment var is optional                                                                           |
| `env:"default=foo"` | denotes the default value if the environment var is missing                                                       |
| `env:"prefix=SUB"`  | _(on a struct field)_ denotes all fields on the embedded struct will load from env var names prefixed with `SUB_` |
| `env:"prefix=SUB_"` | _(on a `map[string]string` field)_ denotes the map will read all env vars whose name starts with `SUB_`           |

## Options
When loading config from environment vars, several option interfaces can be passed to `cfgenv.Load` to alter the names of expected environment vars
or provide support for extra field types.

<details>
    <summary><code>cfgenv.PrefixOption</code></summary>

### `cfgenv.PrefixOption`
Alters the prefix for all environment vars

(Implement interface or use `cfgenv.NewPrefix(prefix string)`

Example:
```go
package main

import (
    "fmt"
    "github.com/go-andiamo/cfgenv"
)

type Config struct {
    ServiceName string
}

func main() {
    cfg := &Config{}
    err := cfgenv.Load(cfg, cfgenv.NewPrefix("MYAPP"))
    if err != nil {
        panic(err)
    } else {
        fmt.Printf("%+v\n", cfg)
    }
}
```
to load from environment variables...
```
MYAPP_SERVICE_NAME=foo
```

</details>
<br>
<details>
    <summary><code>cfgenv.SeparatorOption</code></summary>

### `cfgenv.SeparatorOption`
Alters the separators used between prefixes and field names for environment vars

(Implement interface or use `cfgenv.NewSeparator(separator string)`

Example:
```go
package main

import (
    "fmt"
    "github.com/go-andiamo/cfgenv"
)

type DbConfig struct {
    Host     string `env:"optional,default=localhost"`
    Port     uint   `env:"optional,default=3601"`
    Username string
    Password string
}

type Config struct {
    ServiceName string
    Database    DbConfig `env:"prefix=DB"`
}

func main() {
    cfg := &Config{}
    err := cfgenv.Load(cfg, cfgenv.NewPrefix("MYAPP"), cfgenv.NewSeparator("."))
    if err != nil {
        panic(err)
    } else {
        fmt.Printf("%+v\n", cfg)
    }
}
```
to load from environment variables...
```
MYAPP.SERVICE_NAME=foo
MYAPP.DB.HOST=localhost
MYAPP.DB.PORT=33601
MYAPP.DB.USERNAME=root
MYAPP.DB.PASSWORD=root
```

</details>
<br>
<details>
    <summary><code>cfgenv.NamingOption</code></summary>

### `cfgenv.NamingOption`
Overrides how environment variable names are deduced from field names

Example:
```go
package main

import (
    "fmt"
    "github.com/go-andiamo/cfgenv"
    "reflect"
    "strings"
)

type DbConfig struct {
    Host     string `env:"optional,default=localhost"`
    Port     uint   `env:"optional,default=3601"`
    Username string
    Password string
}

type Config struct {
    ServiceName string
    Database    DbConfig `env:"prefix=DB"`
}

func main() {
    cfg := &Config{}
    err := cfgenv.Load(cfg, &LowercaseFieldNames{}, cfgenv.NewSeparator("."))
    if err != nil {
        panic(err)
    } else {
        fmt.Printf("%+v\n", cfg)
    }
}

type LowercaseFieldNames struct{}

func (l *LowercaseFieldNames) BuildName(prefix string, separator string, fld reflect.StructField, overrideName string) string {
    name := overrideName
    if name == "" {
        name = strings.ToLower(fld.Name)
    }
    if prefix != "" {
        name = prefix + separator + name
    }
    return name
}
```
to load from environment variables...
```
servicename=foo
DB.host=localhost
DB.port=33601
DB.username=root
DB.password=root
```
</details>
<br>
<details>
    <summary><code>cfgenv.CustomSetterOption</code></summary>

### `cfgenv.CustomSetterOption`
Provides support for custom struct field types

Example - see [custom_setter_option](https://github.com/go-andiamo/cfgenv/tree/main/_examples/custom_setter_option)
</details>

## Write Example
Cfgenv can also write examples and current config using the `cfgenv.Example()` or `cfgenv.Write()` functions.

Example - see [write_example](https://github.com/go-andiamo/cfgenv/tree/main/_examples/write_example)
