package internal

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
)

type Kustomizer struct {
	shell   string
	kubectl string
}

func NewKustomizer(shell, kubectl string) *Kustomizer {
	return &Kustomizer{
		shell:   shell,
		kubectl: kubectl,
	}
}

type KustomizeRequest struct {
	Root string
}

type KustomizeResultItem struct {
	KustomizeDir string
	Data         string
}

type KustomizeResult struct {
	Items []*KustomizeResultItem
}

// Render kustomize dir for each kustomization.yaml under req.Root.
func (k Kustomizer) Kustomize(ctx context.Context, req *KustomizeRequest) (*KustomizeResult, error) {
	files, err := FindFiles("kustomization.yaml", req.Root)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to find kustomization.yaml", err)
	}

	const concurrency = 4

	type result struct {
		item *KustomizeResultItem
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
				r := NewCmd(k.shell, fmt.Sprintf("%s kustomize", k.kubectl)).Arg(x).Run(iCtx)
				if err := r.Err; err != nil {
					resultC <- &result{
						err: fmt.Errorf("%w: failed to kustomize %s", err, x),
					}
					cancel()
					return
				}
				resultC <- &result{
					item: &KustomizeResultItem{
						KustomizeDir: x,
						Data:         r.Stdout,
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

	items := make([]*KustomizeResultItem, len(results))
	for i, x := range results {
		if err := x.err; err != nil {
			return nil, err
		}
		items[i] = x.item
	}

	return &KustomizeResult{
		Items: items,
	}, nil
}
