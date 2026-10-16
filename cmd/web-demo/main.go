// web-demo is an isolated server-rendered prototype with process-local changes.
package main

import (
	"flag"
	"github.com/fachschaftinformatik/web/internal/webdemo/account"
	"log"
	"net/http"
	"path/filepath"
	"time"

	web "github.com/fachschaftinformatik/web/internal/webdemo/handler/http"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8090", "HTTP listen address")
	data := flag.String("data", "", "Optional production-copy directory; users are read from sqlite.db without writes")
	flag.Parse()
	users := account.Users("")
	if *data != "" {
		var err error
		users, err = account.LoadUsers(filepath.Join(*data, "sqlite.db"))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Read-only user snapshot: %d accounts", len(users))
	}
	handler, err := web.NewWithUsers(users)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Web demo: http://%s/team", *addr)
	log.Fatal(server.ListenAndServe())
}
