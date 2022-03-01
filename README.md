# go-fang
[![Go Reference](https://pkg.go.dev/badge/github.com/wjiec/go-fang.svg)](https://pkg.go.dev/github.com/wjiec/go-fang)
[![Go Report Card](https://goreportcard.com/badge/github.com/wjiec/go-fang)](https://goreportcard.com/report/github.com/wjiec/go-fang)

Package fang provides a simple and elegant way to bind command line
arguments to struct fields.


### Installation

```bash
go get -u github.com/wjiec/go-fang
```


### Documentation

See [Go Reference](https://pkg.go.dev/github.com/wjiec/go-fang)

More details about fang can be found at [doc.go](doc.go)

### Quick Start

Bind command line arguments to a struct value
```go
package main

import (
    "fmt"
    "net"

    "github.com/spf13/cobra"
    "github.com/wjiec/go-fang"
)

func main() {
    var run struct {
        Config  string  `usage:"Location of client config files" fang:"persistent, required"`
        Context *string `shorthand:"c" usage:"Name of the context to use to connect to the daemon" fang:"p"`
        Debug   bool    `shorthand:"D" usage:"Enable debug mode" fang:"persistent"`
        Host    net.IP  `shorthand:"H" usage:"Daemon socket(s) to connect to"`
        Run     struct {
            Tags         []string          `shorthand:"t" usage:"Set tag data on a container"`
            Environments map[string]string `shorthand:"e" usage:"Set environment variables"`
        }
    }

    root := &cobra.Command{Use: "docker"}
    _ = fang.Bind(root, &run)

    _ = root.ParseFlags([]string{
        "--config", "~/.docker",
        "-c", "abc",
        "-H", "10.0.0.1",
        "-e", "MYSQL_ROOT_PASSWORD=root",
        "-e", "MYSQL_DATABASE=fang",
        "-t", "database",
        "-t", "project-v",
    })

    fmt.Printf("%+v\n", run)
}
```

Bind command line arguments to multi struct value
```go
package main

import (
    "fmt"
    "net"

    "github.com/spf13/cobra"
    "github.com/wjiec/go-fang"
)

func main() {
    var global struct {
        Config  string  `usage:"Location of client config files" fang:"persistent, required"`
        Context *string `shorthand:"c" usage:"Name of the context to use to connect to the daemon" fang:"p"`
        Debug   bool    `shorthand:"D" usage:"Enable debug mode" fang:"persistent"`
        Host    net.IP  `shorthand:"H" usage:"Daemon socket(s) to connect to"`
    }

    var run struct {
        Tags         []string          `shorthand:"t" usage:"Set tag data on a container"`
        Environments map[string]string `shorthand:"e" usage:"Set environment variables"`
    }

    root := &cobra.Command{Use: "docker"}
    b, _ := fang.New(root)

    _ = b.Bind(&global)
    _ = b.Bind(&run)

    _ = root.ParseFlags([]string{
        "--config", "~/.docker",
        "-c", "abc",
        "-H", "10.0.0.1",
        "-e", "MYSQL_ROOT_PASSWORD=root",
        "-e", "MYSQL_DATABASE=fang",
        "-t", "database",
        "-t", "project-v",
    })

    fmt.Printf("global: %+v\nrun: %+v\n", global, run)
}
```

#### Production

At binding time, fields that have been assigned a value will have it as the default value for command line arguments. For example, we first read the configuration from a config file (such as a json or yaml file) and then override the values of those configurations using command line arguments.
```go
package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/wjiec/go-fang"
)

type Config struct {
	Username string `yaml:"username" shorthand:"u" usage:"username to login database"`
	Password string `yaml:"password" usage:"password for username to login"`
}

func main() {
	var cfg Config
	fp, _ := os.Open("config.json")
	content, _ := ioutil.ReadAll(fp)
	_ = json.Unmarshal(content, &cfg)

	root := cobra.Command{Use: "dbc"}
	_ = fang.Bind(&root, &cfg)
	// ...
}
```

It is also possible to set default values for command line arguments by reading environment variables, but note that fang does not provide a way to bind environment variables. This can be done with third-party libraries(e.g. [kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig)).


### License

Released under the [MIT License](LICENSE).
