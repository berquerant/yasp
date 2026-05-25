package internal_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/berquerant/yasp/internal"
	"github.com/stretchr/testify/assert"
)

func TestKustomizer(t *testing.T) {
	root := t.TempDir()
	var (
		k1   = filepath.Join(root, "dir1", "kustomization.yaml")
		k1r1 = filepath.Join(root, "dir1", "resource.yaml")
		k2   = filepath.Join(root, "dir2/dir21/kustomization.yaml")
		k2r1 = filepath.Join(root, "dir2/dir21/resource.yaml")
	)
	for _, x := range []string{k1, k1r1, k2, k2r1} {
		if !assert.Nil(t, os.MkdirAll(filepath.Dir(x), 0755)) {
			return
		}
	}
	for _, x := range []string{k1, k2} {
		if !assert.Nil(t, os.WriteFile(x, []byte(`resources:
- resource.yaml`), 0644)) {
			return
		}
	}
	const resource = `apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx:1.14.2
    name: nginx
    ports:
    - containerPort: 80
`
	for _, x := range []string{k1r1, k2r1} {
		if !assert.Nil(t, os.WriteFile(x, []byte(resource), 0644)) {
			return
		}
	}

	kustomizer := internal.NewKustomizer("bash", "kubectl")
	got, err := kustomizer.Kustomize(context.TODO(), &internal.KustomizeRequest{
		Root: root,
	})
	if !assert.Nil(t, err) {
		return
	}

	want := &internal.KustomizeResult{
		Items: []*internal.KustomizeResultItem{
			{
				KustomizeDir: filepath.Dir(k1),
				Data:         resource,
			},
			{
				KustomizeDir: filepath.Dir(k2),
				Data:         resource,
			},
		},
	}

	sorter := func(x, y *internal.KustomizeResultItem) int {
		switch {
		case x.KustomizeDir > y.KustomizeDir:
			return 1
		case x.KustomizeDir < y.KustomizeDir:
			return -1
		default:
			return 0
		}
	}
	slices.SortStableFunc(got.Items, sorter)
	slices.SortStableFunc(want.Items, sorter)
	assert.Equal(t, want, got)
}
