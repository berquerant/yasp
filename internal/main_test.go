package internal_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/yasp/internal"
	"github.com/berquerant/yasp/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	var (
		fail               = filepath.Join(t.TempDir(), "fail")
		kustomizeRoot      = t.TempDir()
		kustomize1         = filepath.Join(kustomizeRoot, "kustomization.yaml")
		kustomize1resource = filepath.Join(kustomizeRoot, "resource.yaml")
	)
	if !assert.Nil(t, os.WriteFile(fail, []byte(`#!/bin/bash
echo >&2 failed!
exit 1`), 0755)) {
		return
	}
	if !assert.Nil(t, os.WriteFile(kustomize1, []byte(`resources:
- resource.yaml`), 0644)) {
		return
	}
	const resourceYaml = `apiVersion: v1
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
	if !assert.Nil(t, os.WriteFile(kustomize1resource, []byte(resourceYaml), 0644)) {
		return
	}

	for _, tc := range []struct {
		name     string
		args     []string
		stdin    string
		want     string
		exitCode int
	}{
		{
			name: "kustomize",
			args: []string{"--k8s", "cat", "-k", kustomizeRoot, "-o", "verbose"},
			want: fmt.Sprintf(`--- kustomize=%[1]s v1>Pod>>nginx cat
%[2]sdocument.command.cat.count 1, document.command.cat.passed 1, document.count 1, document.process.count 1, document.process.passed 1, kustomize %[1]s
`, kustomizeRoot, resourceYaml),
		},
		{
			name: "k8s id",
			args: []string{"--k8s", "cat"},
			stdin: `---
apiVersion: v1
kind: Pod
metadata:
  namespace: foo
  name: id1
`,
			want: `Summary: 1 document, 0 invalid, 1 processed, 0 failed, 1 passed, 0 denied
`,
		},
		{
			name: "fail success",
			args: []string{"-i", "id", fail, "--success"},
			stdin: `---
id: id1`,
			exitCode: 0,
			want: fmt.Sprintf(`--- id1 %s
failed!
Summary: 1 document, 0 invalid, 1 processed, 0 failed, 0 passed, 1 denied
`, fail),
		},
		{
			name: "fail",
			args: []string{"-i", "id", fail},
			stdin: `---
id: id1`,
			exitCode: 1,
			want: fmt.Sprintf(`--- id1 %s
failed!
Summary: 1 document, 0 invalid, 1 processed, 0 failed, 0 passed, 1 denied
`, fail),
		},
		{
			name: "fail cat",
			args: []string{"-i", "id", fail, "cat"},
			stdin: `---
id: id1`,
			exitCode: 1,
			want: fmt.Sprintf(`--- id1 %s
failed!
Summary: 1 document, 0 invalid, 1 processed, 0 failed, 0 passed, 1 denied
`, fail),
		},
		{
			name: "cat",
			args: []string{"-i", "id", "cat"},
			stdin: `---
id: id1`,
			want: `Summary: 1 document, 0 invalid, 1 processed, 0 failed, 1 passed, 0 denied
`,
		},
		{
			name: "cat verbose",
			args: []string{"-i", "id", "cat", "-o", "verbose"},
			stdin: `---
id: id1`,
			want: `--- id1 cat
id: id1
document.command.cat.count 1, document.command.cat.passed 1, document.count 1, document.process.count 1, document.process.passed 1
`,
		},
		{
			name: "cat yaml",
			args: []string{"-i", "id", "cat", "-o", "yaml"},
			stdin: `---
id: id1`,
			want: `- command: cat
  data: |
    id: id1
  exitcode: 0
  id: id1
  stderr: ""
  stdout: |
    id: id1
- summary:
  - metric: document.command.cat.count
    value: 1
  - metric: document.command.cat.passed
    value: 1
  - metric: document.count
    value: 1
  - metric: document.process.count
    value: 1
  - metric: document.process.passed
    value: 1
`,
		},
		{
			name: "cat verbose 2",
			args: []string{"-i", "id", "cat", "-o", "verbose"},
			stdin: `---
id: id1
---
id: id2`,
			want: `--- id1 cat
id: id1
--- id2 cat
id: id2
document.command.cat.count 2, document.command.cat.passed 2, document.count 2, document.process.count 2, document.process.passed 2
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metric.Reset()
			var stdout bytes.Buffer
			exitCode := internal.Main(context.TODO(), &stdout, bytes.NewBufferString(tc.stdin), tc.args...)
			assert.Equal(t, tc.exitCode, exitCode)
			assert.Equal(t, tc.want, stdout.String())
		})
	}
}
