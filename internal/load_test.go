package internal_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/berquerant/yasp/internal"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		tmpl  string
		want  *internal.DocumentList
		err   error
	}{
		{
			name: "invalid id tmpl",
			input: `---
x: 1`,
			tmpl: `{{`,
			err:  internal.ErrInvalidIDGenerator,
		},
		{
			name:  "invalid input",
			input: `[`,
			tmpl:  `name`,
			err:   internal.ErrYamlUnmarshal,
		},
		{
			name: "load 1 document",
			input: `---
x: 1
name: "id1"`,
			tmpl: `name`,
			want: &internal.DocumentList{
				Items: []*internal.Document{
					{
						ID: "id1",
						Data: `name: id1
x: 1
`,
					},
				},
			},
		},
		{
			name: "load 2 documents but 1 failed",
			input: `---
x: 1
name: "id1"
---
x: 2`,
			tmpl: `name`,
			want: &internal.DocumentList{
				Items: []*internal.Document{
					{
						ID: "id1",
						Data: `name: id1
x: 1
`,
					},
					{
						ID: internal.UnknownDocumentID,
						Data: `x: 2
`,
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := internal.Load(context.TODO(), bytes.NewBufferString(tc.input), tc.tmpl)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			if !assert.Nil(t, err) {
				t.Logf("error=%v", err)
				return
			}
			if !assert.Len(t, got.Items, len(tc.want.Items)) {
				return
			}
			for i, w := range tc.want.Items {
				g := got.Items[i]
				assert.Equal(t, w.ID, g.ID, "index=%d", i)
				assert.Equal(t, w.Data, g.Data, "index=%d", i)
			}
		})
	}
}
