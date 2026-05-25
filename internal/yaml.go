package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
)

type Marshaler interface {
	Marshal(ctx context.Context, v any) ([]byte, error)
}

var _ Marshaler = &YamlMarshaler{}

type Unmarshaler[T any] interface {
	// Unmarshal reads all documents and unmarshal them as a list of T.
	Unmarshal(ctx context.Context) ([]T, error)
}

var _ Unmarshaler[string] = &YamlUnmarshaler[string]{}

type YamlMarshaler struct {
	indent                  int
	literalStyleIfMultiline bool
}

func NewYamlMarshaler(indent int, literalStyleIfMultiline bool) *YamlMarshaler {
	return &YamlMarshaler{
		indent:                  indent,
		literalStyleIfMultiline: literalStyleIfMultiline,
	}
}

var ErrYamlMarshal = errors.New("YamlMarshal")

func (y *YamlMarshaler) Marshal(ctx context.Context, v any) ([]byte, error) {
	b, err := yaml.MarshalContext(
		ctx,
		v,
		yaml.Indent(y.indent),
		yaml.UseLiteralStyleIfMultiline(y.literalStyleIfMultiline),
	)
	if err != nil {
		return nil, errors.Join(ErrYamlMarshal, err)
	}
	return b, nil
}

type YamlUnmarshaler[T any] struct {
	r                    io.Reader
	allowDuplicateMapKey bool
}

func NewYamlUnmarshaler[T any](r io.Reader, allowDuplicateMapKey bool) *YamlUnmarshaler[T] {
	return &YamlUnmarshaler[T]{
		r:                    r,
		allowDuplicateMapKey: allowDuplicateMapKey,
	}
}

var ErrYamlUnmarshal = errors.New("YamlUnmarshal")

func (y *YamlUnmarshaler[T]) Unmarshal(ctx context.Context) ([]T, error) {
	b, err := io.ReadAll(y.r)
	if err != nil {
		return nil, errors.Join(ErrYamlUnmarshal, fmt.Errorf("unmarshal read: %w", err))
	}

	var opts []parser.Option
	if y.allowDuplicateMapKey {
		opts = append(opts, parser.AllowDuplicateMapKey())
	}
	fileNode, err := parser.ParseBytes(b, parser.ParseComments, opts...)
	if err != nil {
		return nil, errors.Join(ErrYamlUnmarshal, fmt.Errorf("unmarshal parse: %w", err))
	}

	decoderOpts := []yaml.DecodeOption{
		yaml.UseOrderedMap(),
	}
	if y.allowDuplicateMapKey {
		decoderOpts = append(decoderOpts, yaml.AllowDuplicateMapKey())
	}
	result := []T{}
	for i, d := range fileNode.Docs {
		if d.Body == nil {
			slog.Info("skip to load document due to empty", slog.Int("index", i))
			continue
		}
		t := new(T)
		decoder := yaml.NewDecoder(strings.NewReader(d.String()), decoderOpts...)
		err := decoder.DecodeFromNode(d.Body, t)
		if err != nil {
			slog.Info("failed to load document", slog.Int("index", i), slog.Any("err", err))
			continue
		}
		result = append(result, *t)
	}
	return result, nil
}
