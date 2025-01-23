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
* load config from environment variables or from file (e.g. `.env` file) or any other `io.Reader`

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

| Tag                                    | Purpose                                                                                                                                                                                                                                                                                 |
|----------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `env:"MY"`                             | overrides the environment var name to read with `MY`                                                                                                                                                                                                                                    |
| `env:"optional"`                       | denotes the environment var is optional                                                                                                                                                                                                                                                 |
| `env:"default=foo"`                    | denotes the default value if the environment var is missing                                                                                                                                                                                                                             |
| `env:"prefix=SUB"`                     | _(on a struct field)_ denotes all fields on the embedded struct will load from env var names prefixed with `SUB_`                                                                                                                                                                       |
| `env:"prefix=SUB_"`                    | _(on a `map[string]string` field)_ denotes the map will read all env vars whose name starts with `SUB_`                                                                                                                                                                                 |
| `env:"match='\d{3}'"`                  | _(on a `map[string]string` field)_ denotes the map will read all env vars whose name matches the regexp `\d{3}`                                                                                                                                                                         |
| `env:"delimiter=;"`<br>`env:"delim=;"` | _(on `slice` and `map` fields)_ denotes the character used to delimit items<br>_(the default is `,`)_                                                                                                                                                                                   |
| `env:"separator=:"`<br>`env:"sep=:"`   | _(on `map` fields)_ denotes the character used to separate key and value<br>_(the default is `:`)_                                                                                                                                                                                      |
| `env:"encodng=base64"`                | denotes the environment var is encoded as `base64` and will be decoded.<br>Built-in decoders are `base64`, `base64url`, `rawBase64` (no padding) & `rawBase64url` (no padding)<br>Other decoders are supported by passing a `Decoder` interface as an option to `Load()`/`LoadAs()` |


## Options
When loading config from environment vars, several option interfaces can be passed to `cfgenv.Load()` function to alter the names of expected environment vars
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
    <summary><code>cfgenv.ExpandOption</code></summary>

### `cfgenv.ExpandOption`
Providing an <code>cfgenv.ExpandOption</code> to the <code>cfgenv.Load()</code> function allows support for resolving substitute environment variables - e.g. `EXAMPLE=${FOO}-{$BAR}` 

<em>Use the <code>Expand()</code> function - or implement your own <code>ExpandOption</code></em>

Example - see [expand_option](https://github.com/go-andiamo/cfgenv/tree/main/_examples/expand_option)

</details>
<br>
<details>
    <summary><code>cfgenv.CustomSetterOption</code></summary>

### `cfgenv.CustomSetterOption`
Provides support for custom struct field types

Example - see [custom_setter_option](https://github.com/go-andiamo/cfgenv/tree/main/_examples/custom_setter_option)
</details>
<br>
<details>
    <summary><code>cfgenv.EnvReader</code></summary>

### `cfgenv.EnvReader`
Reads environment vars from specified reader (e.g. `cfgenv.NewEnvFileReader()`)

Example:

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/cfgenv"
    "os"
)

type Config struct {
    ServiceName string
}

func main() {
    cfg := &Config{}
    f, err := os.Open("local.env")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    err = cfgenv.Load(cfg, cfgenv.NewEnvFileReader(f, nil))
    if err != nil {
        panic(err)
    } else {
        fmt.Printf("%+v\n", cfg)
    }
}
```
where file `local.env` looks like...
```
# this is the service name...
SERVICE_NAME=foo
```

</details>




## Write Example
Cfgenv can also write examples and current config using the `cfgenv.Example()` or `cfgenv.Write()` functions.

Example - see [write_example](https://github.com/go-andiamo/cfgenv/tree/main/_examples/write_example)
