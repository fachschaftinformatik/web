package disk

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestObjectBoundaryAndAtomicUploads(t *testing.T) {
	root := t.TempDir()
	store, err := New(filepath.Join(root, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	if err := store.Upload(ctx, "archive/1/source.pdf", strings.NewReader("original"), 8, "application/pdf"); err != nil {
		t.Fatal(err)
	}
	if err := store.Upload(ctx, "archive/1/source.pdf", strings.NewReader("partial"), 100, "application/pdf"); err == nil {
		t.Fatal("incomplete upload succeeded")
	}
	object, err := store.GetObject(ctx, "archive/1/source.pdf")
	if err != nil {
		t.Fatal(err)
	}
	value, err := io.ReadAll(object)
	object.Close()
	if err != nil || string(value) != "original" {
		t.Fatal("failed upload replaced existing object")
	}
	if err := os.WriteFile(filepath.Join(root, "private"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "private"), filepath.Join(root, "objects", "link")); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"../private", filepath.Join(root, "private"), "link"} {
		if object, err := store.GetObject(ctx, key); err == nil {
			object.Close()
			t.Fatalf("object escaped root: %q", key)
		}
	}
	if err := store.Upload(ctx, "../private", strings.NewReader("overwrite"), 9, "text/plain"); err == nil {
		t.Fatal("upload escaped root")
	}
}
