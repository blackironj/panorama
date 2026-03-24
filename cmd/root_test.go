package cmd

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blackironj/panorama/conv"
)

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
			assert.Equal(t, tt.want, isValidSide(tt.side))
		})
	}
}

func TestResolveTargetSides(t *testing.T) {
	t.Parallel()

	t.Run("empty returns all sides", func(t *testing.T) {
		t.Parallel()
		got, err := resolveTargetSides(nil)
		require.NoError(t, err)
		assert.Equal(t, validSides, got)
	})

	t.Run("valid subset", func(t *testing.T) {
		t.Parallel()
		got, err := resolveTargetSides([]string{"front", "back"})
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("invalid side returns error", func(t *testing.T) {
		t.Parallel()
		_, err := resolveTargetSides([]string{"front", "invalid"})
		assert.Error(t, err)
	})
}

func TestParseInterpolation(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		for _, name := range []string{"nearest", "bilinear", "bicubic"} {
			_, err := conv.ParseInterpolation(name)
			assert.NoError(t, err, name)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		_, err := conv.ParseInterpolation("invalid")
		assert.Error(t, err)
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
			require.NoError(t, err)
			assert.Equal(t, tt.want, isImageFile(fs.FileInfoToDirEntry(entry)))
		})
	}
}
