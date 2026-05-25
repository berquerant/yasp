package internal

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"unicode"

	"github.com/berquerant/yasp/internal/metric"
)

type Runner struct {
	cmdRunner *CmdRunner
	failed    bool
}

func NewRunner(cmdRunner *CmdRunner) *Runner {
	return &Runner{
		cmdRunner: cmdRunner,
	}
}

func (r Runner) Failed() bool { return r.failed }
func (r *Runner) setFailed(reason string) {
	slog.Debug("Runner: set failed", slog.String("reason", reason))
	r.failed = true
}

type RunRequest struct {
	Documents   *DocumentList
	FailFast    bool
	FailFastCmd bool
}

type RunResult struct {
	Result *CmdRunnerResult
}

const (
	runnerMetricDocumentCount         = "document.count"
	runnerMetricDoucmentInvalid       = "document.invalid"
	runnerMetricDocumentProcessCount  = "document.process.count"
	runnerMetricDocumentProcessFailed = "document.process.failed"
	runnerMetricDocumentProcessPassed = "document.process.passed"
	runnerMetricDocumentProcessDenied = "document.process.denied"
)

// Run validations for each yaml document.
func (r *Runner) Run(ctx context.Context, req *RunRequest) chan *RunResult {
	resultC := make(chan *RunResult, 100)

	go func() {
		defer close(resultC)

		for i, doc := range req.Documents.Items {
			metric.Incr(runnerMetricDocumentCount)
			logger := slog.With(slog.Int("index", i), slog.String("id", doc.ID))
			logger.Debug("Check document")
			if doc.Err != nil {
				metric.Incr(runnerMetricDoucmentInvalid)
				logger.Error("Invalid document", slog.String("err", doc.Err.Error()))
				continue
			}

			metric.Incr(runnerMetricDocumentProcessCount)
			logger.Debug("Start to process document", slog.Int("size", len([]byte(doc.Data))))
			result, err := r.cmdRunner.Run(ctx, &CmdRunRequest{
				Document: doc,
				FailFast: req.FailFastCmd,
			})
			if err != nil {
				metric.Incr(runnerMetricDocumentProcessFailed)
				slog.Error("Failed to process document", slog.String("err", err.Error()))
				r.setFailed(fmt.Sprintf("failed to process document id=%s, err=%v", doc.ID, err))
				if req.FailFast {
					return
				}
				continue
			}

			resultC <- &RunResult{
				Result: result,
			}

			for _, x := range result.Items {
				c := x.Command
				if strings.ContainsFunc(c, unicode.IsSpace) {
					c = fmt.Sprintf(`'%s'`, c)
				}
				metric.Incr(fmt.Sprintf("document.command.%s.count", c))
				if x.Result.ExitCode == 0 {
					metric.Incr(fmt.Sprintf("document.command.%s.passed", c))
				} else {
					metric.Incr(fmt.Sprintf("document.command.%s.denied", c))
				}
			}

			if idx := slices.IndexFunc(result.Items, func(x *CmdRunnerResultItem) bool {
				return x.Result.Err != nil
			}); idx >= 0 {
				metric.Incr(runnerMetricDocumentProcessDenied)
				v := result.Items[idx]
				r.setFailed(fmt.Sprintf("failed to run command, document=%s, command=%s, err=%v", doc.ID, v.Command, v.Result.Err))
				if req.FailFast {
					return
				}
			} else {
				metric.Incr(runnerMetricDocumentProcessPassed)
			}

		}
	}()

	return resultC
}
