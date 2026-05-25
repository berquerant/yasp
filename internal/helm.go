package internal

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
)

type HelmRenderer struct {
	shell string
	helm  string
}

func NewHelmRenderer(shell, helm string) *HelmRenderer {
	return &HelmRenderer{
		shell: shell,
		helm:  helm,
	}
}

type HelmRenderRequest struct {
	Root string
}

type HelmRenderResultItem struct {
	ChartDir string
	Data     string
}

type HelmRenderResult struct {
	Items []*HelmRenderResultItem
}

// Render chart for each Chart.yaml under the req.Root.
func (h HelmRenderer) Render(ctx context.Context, req *HelmRenderRequest) (*HelmRenderResult, error) {
	files, err := FindFiles("Chart.yaml", req.Root)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to find Chart.yaml", err)
	}

	const concurrency = 4

	type result struct {
		item *HelmRenderResultItem
		err  error
	}

	var (
		sendC        = make(chan string, concurrency)
		resultC      = make(chan *result, 100)
		wg           sync.WaitGroup
		iCtx, cancel = context.WithCancel(ctx)
	)
	defer cancel()

	wg.Add(concurrency)
	for range concurrency {
		go func() {
			defer wg.Done()
			for x := range sendC {
				c := NewCmd(h.shell, fmt.Sprintf("%s dependency build", h.helm))
				c.Dir = x
				if err := c.Run(iCtx).Err; err != nil {
					resultC <- &result{
						err: fmt.Errorf("%w: failed to build helm dependency for %s", err, x),
					}
					cancel()
					return
				}

				c = NewCmd(h.shell, fmt.Sprintf("%s template yasp .", h.helm))
				c.Dir = x
				r := c.Run(iCtx)
				if err := r.Err; err != nil {
					resultC <- &result{
						err: fmt.Errorf("%w: failed to helm render %s", err, x),
					}
					cancel()
					return
				}
				resultC <- &result{
					item: &HelmRenderResultItem{
						ChartDir: x,
						Data:     r.Stdout,
					},
				}
			}
		}()
	}

	for _, x := range files {
		sendC <- filepath.Dir(x)
	}
	close(sendC)

	var (
		results []*result
		doneC   = make(chan struct{})
	)
	go func() {
		defer close(doneC)
		for x := range resultC {
			results = append(results, x)
		}
	}()

	wg.Wait()
	close(resultC)
	<-doneC

	items := make([]*HelmRenderResultItem, len(results))
	for i, x := range results {
		if err := x.err; err != nil {
			return nil, err
		}
		items[i] = x.item
	}

	return &HelmRenderResult{
		Items: items,
	}, nil
}
