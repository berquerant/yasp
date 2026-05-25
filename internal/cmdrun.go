package internal

import (
	"context"
	"fmt"
	"os"
	"sync"
)

type CmdRunner struct {
	workDir  string
	commands []*Cmd
}

func NewCmdRunner(command ...*Cmd) *CmdRunner {
	return &CmdRunner{
		commands: command,
	}
}

func (r *CmdRunner) WorkDir(d string) {
	r.workDir = d
}

type CmdRunRequest struct {
	Document *Document
	FailFast bool
}

type CmdRunnerResult struct {
	Items []*CmdRunnerResultItem
}

type CmdRunnerResultItem struct {
	Command  string
	Document *Document
	Result   *CmdResult
}

func (r CmdRunner) Run(ctx context.Context, req *CmdRunRequest) (*CmdRunnerResult, error) {
	if err := req.Document.Err; err != nil {
		return nil, err
	}

	tmp, err := os.CreateTemp(r.workDir, "yasp")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create tmpfile for document %s", err, req.Document.ID)
	}
	if r.workDir == "" {
		defer os.Remove(tmp.Name())
	}

	if _, err := fmt.Fprint(tmp, req.Document.Data); err != nil {
		return nil, fmt.Errorf("%w: failed to write document data to tmpfile for %s", err, req.Document.ID)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("%w: failed to close tmpfile for %s", err, req.Document.ID)
	}

	iCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg    sync.WaitGroup
		itemC = make(chan *CmdRunnerResultItem, len(r.commands))
	)
	for _, c := range r.commands {
		wg.Go(func() {
			r := c.Arg(tmp.Name()).Run(iCtx)
			if req.FailFast && r.Err != nil {
				cancel()
			}
			itemC <- &CmdRunnerResultItem{
				Command:  c.command,
				Document: req.Document,
				Result:   r,
			}
		})
	}
	wg.Wait()
	close(itemC)

	var items []*CmdRunnerResultItem
	for x := range itemC {
		items = append(items, x)
	}

	return &CmdRunnerResult{
		Items: items,
	}, nil
}
