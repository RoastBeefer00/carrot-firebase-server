package views

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
)

// assetVersions maps a static file's URL path to a hash of its contents.
var assetVersions = map[string]string{}

// LoadAssetVersions hashes every file under dir, so pages can link to each one
// by a URL that changes whenever the file does. Caches such as Cloudflare then
// fetch new CSS and JS after every deploy instead of serving the old copy.
func LoadAssetVersions(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		assetVersions["/"+filepath.ToSlash(path)] = hex.EncodeToString(sum[:])[:12]
		return nil
	})
}

// Asset returns the versioned URL for a static file, such as
// "/dist/tailwind.css?v=3f2a9c01b7de".
func Asset(path string) string {
	if v, ok := assetVersions[path]; ok {
		return path + "?v=" + v
	}
	return path
}
