package internal_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/berquerant/yasp/internal"
	"github.com/stretchr/testify/assert"
)

func TestHelmRenderer(t *testing.T) {
	var (
		root           = t.TempDir()
		chart1         = filepath.Join(root, "c1", "Chart.yaml")
		chart1resource = filepath.Join(root, "c1", "templates", "resource.yaml")
		chart2         = filepath.Join(root, "c2", "Chart.yaml")
		chart2resource = filepath.Join(root, "c2", "templates", "resource.yaml")
	)
	const chart = `apiVersion: v2
name: yasp
type: application
version: "0.0.1"
appVersion: "0.0.1"
`
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
	for _, x := range []string{chart1, chart1resource, chart2, chart2resource} {
		if !assert.Nil(t, os.MkdirAll(filepath.Dir(x), 0755)) {
			return
		}
	}
	for _, x := range []string{chart1, chart2} {
		if !assert.Nil(t, os.WriteFile(x, []byte(chart), 0644)) {
			return
		}
	}
	for _, x := range []string{chart1resource, chart2resource} {
		if !assert.Nil(t, os.WriteFile(x, []byte(resource), 0644)) {
			return
		}
	}

	renderer := internal.NewHelmRenderer("bash", "helm")
	got, err := renderer.Render(context.TODO(), &internal.HelmRenderRequest{
		Root: root,
	})
	if !assert.Nil(t, err) {
		return
	}

	wantData := fmt.Sprintf(`---
# Source: yasp/templates/resource.yaml
%s`, resource)
	want := &internal.HelmRenderResult{
		Items: []*internal.HelmRenderResultItem{
			{
				ChartDir: filepath.Dir(chart1),
				Data:     wantData,
			},
			{
				ChartDir: filepath.Dir(chart2),
				Data:     wantData,
			},
		},
	}

	sorter := func(x, y *internal.HelmRenderResultItem) int {
		switch {
		case x.ChartDir > y.ChartDir:
			return 1
		case x.ChartDir < y.ChartDir:
			return -1
		default:
			return 0
		}
	}
	slices.SortStableFunc(got.Items, sorter)
	slices.SortStableFunc(want.Items, sorter)

	assert.Equal(t, want, got)
}
