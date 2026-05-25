package internal_test

import (
	"encoding/json"
	"testing"

	"github.com/berquerant/yasp/internal"
	"github.com/stretchr/testify/assert"
)

func TestIDGenerator(t *testing.T) {
	t.Run("k8s", func(t *testing.T) {
		const s = `{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "id1"
  }
}`
		m := map[string]any{}
		if err := json.Unmarshal([]byte(s), &m); !assert.Nil(t, err) {
			return
		}
		g, err := internal.NewIDGenerator(`apiVersion + ">" + kind + ">" + (metadata.namespace ?? "") + ">" + metadata.name`)
		if !assert.Nil(t, err) {
			return
		}
		got, err := g.GenerateID(m)
		if !assert.Nil(t, err) {
			return
		}
		assert.Equal(t, "v1>Pod>>id1", got)
	})

	for _, tc := range []struct {
		name string
		s    string
		m    map[string]any
		want string
		err  error
	}{
		{
			name: "no data",
			s:    "id",
			m:    map[string]any{},
			err:  internal.ErrIDGenerationFailure,
		},
		{
			name: "id",
			s:    "id",
			m: map[string]any{
				"id": "id1",
			},
			want: "id1",
		},
		{
			name: "k8s namespaced",
			s:    `apiVersion + ">" + kind + ">" + (metadata.namespace ?? "") + ">" + metadata.name`,
			m: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name":      "id1",
					"namespace": "default",
				},
			},
			want: "v1>Pod>default>id1",
		},
		{
			name: "k8s",
			s:    `apiVersion + ">" + kind + ">" + (metadata.namespace ?? "") + ">" + metadata.name`,
			m: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name": "id1",
				},
			},
			want: "v1>Pod>>id1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, err := internal.NewIDGenerator(tc.s)
			if !assert.Nil(t, err) {
				return
			}
			got, err := g.GenerateID(tc.m)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			if !assert.Nil(t, err) {
				return
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
