package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/controller"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/gallery"
	"github.com/fachschaftinformatik/web/internal/model/meetings"
	"github.com/fachschaftinformatik/web/internal/model/organization"
	"github.com/fachschaftinformatik/web/internal/model/people"
	"github.com/fachschaftinformatik/web/internal/model/users"
	"github.com/fachschaftinformatik/web/internal/storage"
	"github.com/fachschaftinformatik/web/internal/storage/disk"
	"github.com/fachschaftinformatik/web/internal/storage/s3"
)

type application struct {
	handler http.Handler
	db      *sql.DB
	local   *disk.Store
	cancel  context.CancelFunc
	done    chan struct{}
}

func newApplication(ctx context.Context, databasePath, blobPath string) (_ *application, err error) {
	cfg := config.New()
	if err := cfg.ValidateDomains(); err != nil {
		return nil, err
	}
	origin, err := url.Parse(cfg.Domain)
	if err != nil || origin.Host == "" || (origin.Scheme != "http" && origin.Scheme != "https") || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" {
		return nil, errors.New("DOMAIN must be an absolute HTTP or HTTPS URL")
	}
	if err := os.MkdirAll(filepath.Dir(databasePath), 0700); err != nil {
		return nil, err
	}
	if err := database.Migrate(databasePath, log.Default()); err != nil {
		return nil, err
	}
	db, err := database.NewConnection(databasePath)
	if err != nil {
		return nil, err
	}
	app := &application{db: db}
	defer func() {
		if err != nil {
			err = errors.Join(err, app.Close())
		}
	}()
	var objects storage.Provider
	if blobPath != "" || cfg.S3Endpoint == "" {
		if blobPath == "" {
			blobPath = filepath.Join(filepath.Dir(databasePath), "blobs")
		}
		app.local, err = disk.New(blobPath)
		if err != nil {
			return nil, err
		}
		objects = app.local
	} else {
		objects, err = s3.NewClient(cfg)
		if err != nil {
			return nil, err
		}
	}
	team, err := people.Open(ctx, db)
	if err != nil {
		return nil, err
	}
	publications, err := meetings.NewStore(db, objects)
	if err != nil {
		return nil, err
	}
	org, err := organization.NewStore(ctx, db)
	if err != nil {
		return nil, err
	}
	identity := users.NewService(db, cfg.Domain, users.DomainPolicy{Registration: cfg.RegistrationDomains, Accounts: cfg.AccountDomains})
	files := archive.NewService(db, objects)
	app.handler, err = controller.New(controller.Services{
		Users: identity, Archive: files, People: team, Meetings: publications, Organization: org,
		Forum: forum.NewService(db), Gallery: gallery.NewService(db, objects), Events: events.NewService(db, objects),
	}, origin.Scheme == "https")
	if err != nil {
		return nil, err
	}
	maintenanceCtx, cancel := context.WithCancel(ctx)
	app.cancel, app.done = cancel, make(chan struct{})
	go func() {
		defer close(app.done)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		sender := email.NewSender(cfg, org)
		for {
			if err := files.CleanupBlobs(maintenanceCtx); err != nil && maintenanceCtx.Err() == nil {
				log.Printf("Archive file maintenance: %v", err)
			}
			if err := publications.CleanupBlobs(maintenanceCtx); err != nil && maintenanceCtx.Err() == nil {
				log.Printf("Meeting file maintenance: %v", err)
			}
			var err error
			if cfg.SMTPHost != "" {
				err = identity.DeliverMail(maintenanceCtx, sender)
			} else {
				err = identity.PurgeExpired(maintenanceCtx)
			}
			if err != nil && maintenanceCtx.Err() == nil {
				log.Printf("Account maintenance: %v", err)
			}
			select {
			case <-maintenanceCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return app, nil
}

func (a *application) Close() error {
	if a.cancel != nil {
		a.cancel()
		<-a.done
	}
	var err error
	if a.local != nil {
		err = a.local.Close()
	}
	return errors.Join(err, a.db.Close())
}
