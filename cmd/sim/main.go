// Command sim serves the simulator via HTTP.
package main

import (
	"embed"
	"net/http"
)

//go:generate bash -c ""
//go:generate bash ./build-wasm.sh

func main() {
	_ = http.ListenAndServe("localhost:8042", server())
}

//go:embed *.html *.js *.wasm *.css
var contentFS embed.FS

func server() http.Handler { return http.FileServer(http.FS(contentFS)) }
