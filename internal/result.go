package internal

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/berquerant/yasp/internal/metric"
)

type Summary struct {
	d    map[string]uint64
	meta *Map
}

func NewSummary() *Summary {
	return &Summary{
		d:    metric.Read(),
		meta: NewMap(nil),
	}
}

func (s *Summary) Meta(meta *Map) *Summary {
	s.meta = meta
	return s
}

func (s *Summary) Write(ctx context.Context, w io.Writer, mode OutputMode) error {
	switch mode {
	case OutputModeText:
		return s.writeText(w)
	case OutputModeVerbose:
		return s.writeVerbose(w)
	case OutputModeYaml:
		return s.writeYaml(ctx, w)
	default:
		return fmt.Errorf("%w: Unknown output mode", ErrInvalidConfig)
	}
}

func (s *Summary) writeText(w io.Writer) error {
	meta := s.meta.SortedJoinedString("=", ",")
	if meta != "" {
		meta = "(" + meta + ")"
	}

	_, err := fmt.Fprintf(w, "Summary%s: %d document, %d invalid, %d processed, %d failed, %d passed, %d denied\n",
		meta,
		s.d[runnerMetricDocumentCount],
		s.d[runnerMetricDoucmentInvalid],
		s.d[runnerMetricDocumentProcessCount],
		s.d[runnerMetricDocumentProcessFailed],
		s.d[runnerMetricDocumentProcessPassed],
		s.d[runnerMetricDocumentProcessDenied],
	)
	return err
}

func (s *Summary) writeVerbose(w io.Writer) error {
	d := s.meta.Clone()
	for k, v := range s.d {
		d.Set(k, v)
	}
	_, err := fmt.Fprintln(w, d.SortedJoinedString(" ", ", "))
	return err
}

func (s *Summary) writeYaml(ctx context.Context, w io.Writer) error {
	dd := s.meta.Clone()
	for k, v := range s.d {
		dd.Set(k, v)
	}

	d := make([]map[string]any, dd.Len())
	var i int
	for k, v := range dd.SortedValues() {
		d[i] = map[string]any{
			"metric": k,
			"value":  v,
		}
		i++
	}
	b, err := NewYamlMarshaler(2, true).Marshal(ctx, []map[string]any{
		{
			"summary": d,
		},
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s", b)
	return err
}

type Result struct {
	result *RunResult
	meta   *Map
}

func NewResult(result *RunResult) *Result {
	return &Result{
		result: result,
		meta:   NewMap(nil),
	}
}

func (r *Result) Meta(meta *Map) *Result {
	r.meta = meta
	return r
}

func (r *Result) Write(ctx context.Context, w io.Writer, mode OutputMode) error {
	switch mode {
	case OutputModeText:
		return r.writeText(w)
	case OutputModeVerbose:
		return r.writeVerbose(w)
	case OutputModeYaml:
		return r.writeYaml(ctx, w)
	default:
		return fmt.Errorf("%w: Unknown output mode", ErrInvalidConfig)
	}
}

func (r *Result) textHeader(docID, command string) string {
	meta := r.meta.SortedJoinedString("=", " ")
	if meta != "" {
		meta = meta + " "
	}
	return fmt.Sprintf("--- %s%s %s\n", meta, docID, command)
}

func (r *Result) writeText(w io.Writer) error {
	var (
		b     strings.Builder
		write = func(msg string, v ...any) {
			b.WriteString(fmt.Sprintf(msg, v...))
		}
	)
	for _, item := range r.result.Result.Items {
		if item.Result.ExitCode == 0 {
			continue
		}
		b.WriteString(r.textHeader(item.Document.ID, item.Command))
		write("%s", item.Result.Stdout)
		write("%s", item.Result.Stderr)
	}
	_, err := fmt.Fprint(w, b.String())
	return err
}

func (r *Result) writeVerbose(w io.Writer) error {
	var (
		b     strings.Builder
		write = func(msg string, v ...any) {
			b.WriteString(fmt.Sprintf(msg, v...))
		}
	)
	for _, item := range r.result.Result.Items {
		b.WriteString(r.textHeader(item.Document.ID, item.Command))
		write("%s", item.Result.Stdout)
		write("%s", item.Result.Stderr)
	}
	_, err := fmt.Fprint(w, b.String())
	return err
}

func (r *Result) writeYaml(ctx context.Context, w io.Writer) error {
	result := make([]map[string]any, len(r.result.Result.Items))
	for i, item := range r.result.Result.Items {
		result[i] = map[string]any{
			"command":  item.Command,
			"id":       item.Document.ID,
			"data":     item.Document.Data,
			"exitcode": item.Result.ExitCode,
			"stdout":   item.Result.Stdout,
			"stderr":   item.Result.Stderr,
		}
		if err := item.Result.Err; err != nil {
			result[i]["err"] = err.Error()
		}
		for k, v := range r.meta.SortedValues() {
			result[i][k] = v
		}
	}
	b, err := NewYamlMarshaler(2, true).Marshal(ctx, result)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s", b)
	return err
}
