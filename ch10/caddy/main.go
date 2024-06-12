package main

import (
    // Import command module (it has Main function to start caddy server)
    cmd "github.com/caddyserver/caddy/v2/cmd"
    // import standard module of caddy server
    _ "github.com/caddyserver/caddy/v2/modules/standard"
    // import custom module
    _ "github.com/huGgW/network-study-with-go/ch10/caddy-toml-adapter"
    _ "github.com/huGgW/network-study-with-go/ch10/caddy-restrict-prefix"
)

func main() {
    cmd.Main()
}
