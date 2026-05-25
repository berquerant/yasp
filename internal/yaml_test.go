package internal_test

import (
	"context"
	"strings"
	"testing"

	"github.com/berquerant/yasp/internal"
	"github.com/stretchr/testify/assert"
)

func TestYamlUnmarshaler(t *testing.T) {
	type Want struct {
		I int    `yaml:"i"`
		S string `yaml:"s"`
	}

	t.Run("allowDuplicateMapKey", func(t *testing.T) {
		const text = `---
i: 1
i: 1
s: "str"`
		t.Run("true", func(t *testing.T) {
			got, err := internal.NewYamlUnmarshaler[Want](strings.NewReader(text), true).Unmarshal(context.TODO())
			if !assert.Nil(t, err) {
				return
			}
			if !assert.Len(t, got, 1) {
				return
			}
			assert.Equal(t, Want{
				I: 1,
				S: "str",
			}, got[0])
		})
		t.Run("false", func(t *testing.T) {
			_, err := internal.NewYamlUnmarshaler[Want](strings.NewReader(text), false).Unmarshal(context.TODO())
			assert.NotNil(t, err)
		})
	})

	for _, tc := range []struct {
		title string
		text  string
		want  []Want
	}{
		{
			title: "ignore empty",
			text:  ``,
			want:  []Want{},
		},
		{
			title: "ignore empty2",
			text: `---
`,
			want: []Want{},
		},
		{
			title: "ignore comment only",
			text: `---
# comment`,
			want: []Want{},
		},
		{
			title: "ignore empties",
			text: `
---
# comment
---
# comment2
---
---`,
			want: []Want{},
		},
		{
			title: "an element",
			text: `
i: 1
s: "str"`,
			want: []Want{
				{
					I: 1,
					S: "str",
				},
			},
		},
		{
			title: "docs",
			text: `
i: 1
s: "str"
---
i: 2
s: "str2"`,
			want: []Want{
				{
					I: 1,
					S: "str",
				},
				{
					I: 2,
					S: "str2",
				},
			},
		},
		{
			title: "docs with empty",
			text: `
i: 1
s: "str"
---
# empty
---
i: 2
s: "str2"`,
			want: []Want{
				{
					I: 1,
					S: "str",
				},
				{
					I: 2,
					S: "str2",
				},
			},
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			got, err := internal.NewYamlUnmarshaler[Want](
				strings.NewReader(tc.text),
				true,
			).Unmarshal(context.TODO())
			if !assert.Nil(t, err) {
				return
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestYamlMarshaler(t *testing.T) {
	for _, tc := range []struct {
		title     string
		marshaler internal.Marshaler
		obj       any
		want      string
	}{
		{
			title:     "scalar",
			marshaler: internal.NewYamlMarshaler(2, true),
			obj:       1,
			want: `1
`,
		},
		{
			title:     "dict",
			marshaler: internal.NewYamlMarshaler(2, true),
			obj: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
			want: `k1: v1
k2: v2
`,
		},
		{
			title:     "array",
			marshaler: internal.NewYamlMarshaler(2, true),
			obj: []string{
				"k1",
				"k2",
			},
			want: `- k1
- k2
`,
		},
		{
			title:     "fuzz",
			marshaler: internal.NewYamlMarshaler(2, true),
			obj: []any{
				"k1",
				map[string]any{
					"k2": "v2",
					"k3": map[string]any{
						"k4": "v4",
					},
				},
			},
			want: `- k1
- k2: v2
  k3:
    k4: v4
`,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			got, err := tc.marshaler.Marshal(context.TODO(), tc.obj)
			if !assert.Nil(t, err) {
				return
			}
			assert.Equal(t, tc.want, string(got))
		})
	}
}
