package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
)

func Main(ctx context.Context, stdout io.Writer, stdin io.Reader, args ...string) int {
	c, err := parseFlags(args...)
	if err != nil {
		if errors.Is(err, ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	b, _ := json.Marshal(c)
	slog.Debug("Use config", slog.String("config", string(b)))

	r := &mainRunner{
		c: c,
	}

	if err := r.main(ctx, stdout, stdin); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	return 0
}

var ErrRunnerFailed = errors.New("Failed")

type mainRunner struct {
	c *Config
}

func (m *mainRunner) main(ctx context.Context, stdout io.Writer, stdin io.Reader) error {
	if m.c.HelmRoot != "" {
		return m.runHelm(ctx, stdout)
	}
	if m.c.KustomizeRoot != "" {
		return m.runKustomize(ctx, stdout)
	}
	return m.runMain(ctx, stdout, stdin, NewMap(nil))
}

func (m *mainRunner) runHelm(ctx context.Context, stdout io.Writer) error {
	renderer := NewHelmRenderer(m.c.Shell, m.c.Helm)
	hs, err := renderer.Render(ctx, &HelmRenderRequest{
		Root: m.c.HelmRoot,
	})
	if err != nil {
		return err
	}

	var errs []error
	for _, h := range hs.Items {
		meta := NewMap(map[string]any{
			"chart": h.ChartDir,
		})
		stdin := bytes.NewBufferString(h.Data)
		if err := m.runMain(ctx, stdout, stdin, meta); err != nil {
			errs = append(errs, fmt.Errorf("%w: chart=%s", err, h.ChartDir))
		}
	}

	return errors.Join(errs...)
}

func (m *mainRunner) runKustomize(ctx context.Context, stdout io.Writer) error {
	kustomizer := NewKustomizer(m.c.Shell, m.c.Kubectl)
	ks, err := kustomizer.Kustomize(ctx, &KustomizeRequest{
		Root: m.c.KustomizeRoot,
	})
	if err != nil {
		return err
	}

	var errs []error
	for _, k := range ks.Items {
		meta := NewMap(map[string]any{
			"kustomize": k.KustomizeDir,
		})
		stdin := bytes.NewBufferString(k.Data)
		if err := m.runMain(ctx, stdout, stdin, meta); err != nil {
			errs = append(errs, fmt.Errorf("%w: kustomize=%s", err, k.KustomizeDir))
		}
	}

	return errors.Join(errs...)
}

func (m *mainRunner) load(ctx context.Context, stdin io.Reader, meta *Map) (*DocumentList, error) {
	if m.c.Bulk {
		dataBytes, err := io.ReadAll(stdin)
		if err != nil {
			return nil, err
		}
		data := string(dataBytes)
		if v, ok := meta.Get("kustomize"); ok {
			return &DocumentList{
				Items: []*Document{
					{
						ID:   v.(string),
						Data: data,
					},
				},
			}, nil
		}
		if v, ok := meta.Get("chart"); ok {
			return &DocumentList{
				Items: []*Document{
					{
						ID:   v.(string),
						Data: data,
					},
				},
			}, nil
		}
		return &DocumentList{
			Items: []*Document{
				{
					ID:   "bulk",
					Data: data,
				},
			},
		}, nil
	}

	return Load(ctx, stdin, m.c.IDTemplate)
}

func (m *mainRunner) runMain(ctx context.Context, stdout io.Writer, stdin io.Reader, meta *Map) error {
	documents, err := m.load(ctx, stdin, meta)
	if err != nil {
		return err
	}

	cmds := make([]*Cmd, len(m.c.Cmds))
	for i, x := range m.c.Cmds {
		cmds[i] = NewCmd(m.c.Shell, x)
	}
	cmdRunner := NewCmdRunner(cmds...)
	cmdRunner.WorkDir(m.c.WorkDir)
	runner := NewRunner(cmdRunner)

	resultC := runner.Run(ctx, &RunRequest{
		Documents:   documents,
		FailFast:    m.c.FailFast,
		FailFastCmd: m.c.FailFastCmd,
	})
	for r := range resultC {
		if err := NewResult(r).Meta(meta).Write(ctx, stdout, m.c.Output); err != nil {
			return err
		}
	}

	if err := NewSummary().Meta(meta).Write(ctx, stdout, m.c.Output); err != nil {
		return err
	}

	if m.c.Success {
		return nil
	}
	if runner.Failed() {
		return ErrRunnerFailed
	}
	return nil
}
