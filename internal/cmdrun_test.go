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

func TestCmdRunner(t *testing.T) {
	var (
		cat   = filepath.Join(t.TempDir(), "cat")
		hello = filepath.Join(t.TempDir(), "hello")
		fail  = filepath.Join(t.TempDir(), "fail")
	)
	if !assert.Nil(t, os.WriteFile(cat, []byte(`#!/bin/bash
cat "$1"`), 0755)) {
		return
	}
	if !assert.Nil(t, os.WriteFile(hello, []byte(`#!/bin/bash
echo hello
echo >&2 hello`), 0755)) {
		return
	}
	if !assert.Nil(t, os.WriteFile(fail, []byte(`#!/bin/bash
exit 1`), 0755)) {
		return
	}

	for _, tc := range []struct {
		name string
		cmds []*internal.Cmd
		req  *internal.CmdRunRequest
		want *internal.CmdRunnerResult
		err  error
	}{
		{
			name: "cat hello",
			cmds: []*internal.Cmd{
				internal.NewCmd("bash", cat),
				internal.NewCmd("bash", hello),
			},
			req: &internal.CmdRunRequest{
				Document: &internal.Document{
					ID:   "id1",
					Data: `data`,
				},
			},
			want: &internal.CmdRunnerResult{
				Items: []*internal.CmdRunnerResultItem{
					{
						Command: hello,
						Document: &internal.Document{
							ID:   "id1",
							Data: `data`,
						},
						Result: &internal.CmdResult{
							Stdout: "hello\n",
							Stderr: "hello\n",
						},
					},
					{
						Command: cat,
						Document: &internal.Document{
							ID:   "id1",
							Data: `data`,
						},
						Result: &internal.CmdResult{
							Stdout: "data",
						},
					},
				},
			},
		},
		{
			name: "cat",
			cmds: []*internal.Cmd{internal.NewCmd("bash", cat)},
			req: &internal.CmdRunRequest{
				Document: &internal.Document{
					ID:   "id1",
					Data: `data`,
				},
			},
			want: &internal.CmdRunnerResult{
				Items: []*internal.CmdRunnerResultItem{
					{
						Command: cat,
						Document: &internal.Document{
							ID:   "id1",
							Data: `data`,
						},
						Result: &internal.CmdResult{
							Stdout: "data",
						},
					},
				},
			},
		},
		{
			name: "fail",
			cmds: []*internal.Cmd{internal.NewCmd("bash", fail)},
			req: &internal.CmdRunRequest{
				Document: &internal.Document{
					ID:   "id1",
					Data: `data`,
				},
			},
			want: &internal.CmdRunnerResult{
				Items: []*internal.CmdRunnerResultItem{
					{
						Command: fail,
						Document: &internal.Document{
							ID:   "id1",
							Data: `data`,
						},
						Result: &internal.CmdResult{
							ExitCode: 1,
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := internal.NewCmdRunner(tc.cmds...).Run(context.TODO(), tc.req)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			if !assert.Nil(t, err) {
				return
			}
			sorter := func(x, y *internal.CmdRunnerResultItem) int {
				switch {
				case x.Command > y.Command:
					return 1
				case x.Command < y.Command:
					return -1
				default:
					return 0
				}
			}
			slices.SortStableFunc(tc.want.Items, sorter)
			slices.SortStableFunc(got.Items, sorter)
			for _, x := range got.Items {
				x.Result.Err = nil // ignore error
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
