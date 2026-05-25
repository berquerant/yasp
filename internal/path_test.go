package internal_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/berquerant/yasp/internal"
	"github.com/stretchr/testify/assert"
)

func TestFindFiles(t *testing.T) {
	root := t.TempDir()
	// root/
	//   dir1/
	//     kustomization.yaml
	//   dir2/
	//     some.txt
	//   dir3/
	//     dir31/
	//       kustomization.yaml
	//   kustomization.yaml
	for _, x := range []string{
		"dir1/kustomization.yaml",
		"dir2/some.txt",
		"dir3/dir31/kustomization.yaml",
		"kustomization.yaml",
	} {
		p := filepath.Join(root, x)
		t.Logf("create %s", p)
		if !assert.Nil(t, os.MkdirAll(filepath.Dir(p), 0755)) {
			return
		}
		if !assert.Nil(t, os.WriteFile(p, []byte(""), 0644)) {
			return
		}
	}

	want := []string{
		filepath.Join(root, "dir1/kustomization.yaml"),
		filepath.Join(root, "dir3/dir31/kustomization.yaml"),
		filepath.Join(root, "kustomization.yaml"),
	}
	got, err := internal.FindFiles("kustomization.yaml", root)
	if !assert.Nil(t, err) {
		return
	}

	slices.Sort(want)
	slices.Sort(got)
	assert.Equal(t, want, got)
}
