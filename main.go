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
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/content/files"
	"github.com/fachschaftinformatik/web/internal/content/media"
	"github.com/fachschaftinformatik/web/internal/content/posts"
	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/data/blob"
	"github.com/fachschaftinformatik/web/internal/data/blob/disk"
	"github.com/fachschaftinformatik/web/internal/data/blob/s3"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/fachschaftinformatik/web/internal/people"
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
	snapshot := flags.String("data", "", "Read-only user snapshot directory for the demo")
	database := flags.String("database", "data/sqlite.db", "Writable SQLite database path")
	blobs := flags.String("blobs", "", "Local object directory; otherwise use configured S3 or blobs beside the database")
	demo := flags.Bool("demo", false, "Use example data and local admin/admin credentials")
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
	var handler http.Handler
	var err error
	if *demo || *snapshot != "" {
		records := users.Examples()
		if *snapshot != "" {
			records, err = users.Load(filepath.Join(*snapshot, "sqlite.db"))
			if err != nil {
				return fmt.Errorf("load account snapshot: %w", err)
			}
		}
		handler, err = web.NewWithUsers(records)
	} else {
		if err := os.MkdirAll(filepath.Dir(*database), 0700); err != nil {
			return err
		}
		if err := data.Migrate(*database, log.Default()); err != nil {
			return err
		}
		db, openErr := data.NewConnection(*database)
		if openErr != nil {
			return openErr
		}
		defer db.Close()
		cfg := config.New()
		origin, parseErr := url.Parse(cfg.Domain)
		if parseErr != nil || origin.Host == "" || (origin.Scheme != "http" && origin.Scheme != "https") || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" {
			return errors.New("DOMAIN must be an absolute HTTP or HTTPS URL")
		}
		identity := users.NewService(db, cfg.Domain)
		var storage blob.Provider
		if *blobs != "" || cfg.S3Endpoint == "" {
			path := *blobs
			if path == "" {
				path = filepath.Join(filepath.Dir(*database), "blobs")
			}
			local, err := disk.New(path)
			if err != nil {
				return err
			}
			defer local.Close()
			storage = local
		} else if cfg.S3Endpoint != "" {
			client, err := s3.NewClient(cfg)
			if err != nil {
				return err
			}
			storage = client
		}
		var team *people.Store
		team, err = people.Open(ctx, db)
		if err != nil {
			return err
		}
		meetings, meetingErr := content.NewPersistentStore(db, storage)
		if meetingErr != nil {
			return meetingErr
		}
		handler, err = web.NewPersistent(web.Services{Users: identity, Archive: files.NewService(db, storage), People: team, Meetings: meetings, Posts: posts.NewService(db), Media: media.NewService(db, storage)}, origin.Scheme == "https")
		if err == nil {
			mailCtx, cancel := context.WithCancel(ctx)
			done := make(chan struct{})
			go func() {
				defer close(done)
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				sender := email.NewSender(cfg)
				for {
					if err := meetings.CleanupBlobs(mailCtx); err != nil && mailCtx.Err() == nil {
						log.Printf("Meeting file maintenance: %v", err)
					}
					var err error
					if cfg.SMTPHost != "" {
						err = identity.DeliverMail(mailCtx, sender)
					} else {
						err = identity.PurgeExpired(mailCtx)
					}
					if err != nil && mailCtx.Err() == nil {
						log.Printf("Account maintenance: %v", err)
					}
					select {
					case <-mailCtx.Done():
						return
					case <-ticker.C:
					}
				}
			}()
			defer func() { cancel(); <-done }()
		}
	}
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
