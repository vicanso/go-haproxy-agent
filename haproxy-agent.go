package main

import . "github.com/tj/go-debug"
import (
	"./consul"
	"path"
)

var debug = Debug("jt.haproxy-agent")

func main() {
	debug("url %s", path.Join("http://192.168.1.1", "/test", "abc"))
	consul.HttpBackends()
}
