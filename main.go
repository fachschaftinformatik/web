// web is the server-rendered Fachschaft website.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fachschaftinformatik/web/internal/users"
	"github.com/fachschaftinformatik/web/internal/web"
)

// Version is injected by the image build.
var Version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("web", flag.ContinueOnError)
	addr := flags.String("addr", "127.0.0.1:8090", "HTTP listen address")
	data := flags.String("data", "", "Optional read-only snapshot directory containing sqlite.db")
	version := flags.Bool("version", false, "Print application version")
	health := flags.String("check-health", "", "Check a running application's health URL and exit")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if *version {
		fmt.Println(Version)
		return nil
	}
	if *health != "" {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get(*health)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("health check returned HTTP %d", response.StatusCode)
		}
		return nil
	}
	records := users.Examples()
	if *data != "" {
		var err error
		records, err = users.Load(filepath.Join(*data, "sqlite.db"))
		if err != nil {
			return fmt.Errorf("load account snapshot: %w", err)
		}
		log.Printf("Read-only account snapshot: %d accounts", len(records))
	}
	handler, err := web.NewWithUsers(records)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	return serve(ctx, listener, handler)
}

func serve(ctx context.Context, listener net.Listener, handler http.Handler) error {
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 120 * time.Second}
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	log.Printf("Web %s: http://%s/", Version, listener.Addr())
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return err
		}
		if err := <-result; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
