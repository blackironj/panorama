package cmd

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/blackironj/panorama/conv"
)

func TestParseInterpolation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"nearest", false},
		{"bilinear", false},
		{"bicubic", false},
		{"invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := conv.ParseInterpolation(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInterpolation(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidSide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		side string
		want bool
	}{
		{"front", true},
		{"back", true},
		{"left", true},
		{"right", true},
		{"top", true},
		{"bottom", true},
		{"invalid", false},
		{"Front", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.side, func(t *testing.T) {
			t.Parallel()
			if got := isValidSide(tt.side); got != tt.want {
				t.Errorf("isValidSide(%q) = %v, want %v", tt.side, got, tt.want)
			}
		})
	}
}

func TestResolveTargetSides(t *testing.T) {
	t.Parallel()

	t.Run("empty returns all sides", func(t *testing.T) {
		t.Parallel()
		got, err := resolveTargetSides(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(validSides) {
			t.Errorf("expected %d sides, got %d", len(validSides), len(got))
		}
	})

	t.Run("valid subset", func(t *testing.T) {
		t.Parallel()
		input := []string{"front", "back"}
		got, err := resolveTargetSides(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("expected 2 sides, got %d", len(got))
		}
	})

	t.Run("invalid side returns error", func(t *testing.T) {
		t.Parallel()
		_, err := resolveTargetSides([]string{"front", "invalid"})
		if err == nil {
			t.Fatal("expected error for invalid side, got nil")
		}
	})
}

func TestIsImageFile(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"photo.jpg":  &fstest.MapFile{},
		"photo.jpeg": &fstest.MapFile{},
		"photo.png":  &fstest.MapFile{},
		"photo.JPG":  &fstest.MapFile{},
		"doc.txt":    &fstest.MapFile{},
		"data.gif":   &fstest.MapFile{},
		"noext":      &fstest.MapFile{},
	}

	tests := []struct {
		name string
		want bool
	}{
		{"photo.jpg", true},
		{"photo.jpeg", true},
		{"photo.png", true},
		{"photo.JPG", true},
		{"doc.txt", false},
		{"data.gif", false},
		{"noext", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry, err := fs.Stat(fsys, tt.name)
			if err != nil {
				t.Fatal(err)
			}
			got := isImageFile(fs.FileInfoToDirEntry(entry))
			if got != tt.want {
				t.Errorf("isImageFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
