// Package jobs executes jobs assigned by the controller and streams output.
package jobs

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/managerhub/managerhub/shared/protocol"
)

// Emitter sends protocol envelopes back to the controller.
type Emitter interface {
	Send(env protocol.Envelope)
}

// Runner manages running job processes.
type Runner struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	emit    Emitter
	nodeID  string
}

// NewRunner creates a job runner.
func NewRunner(emit Emitter, nodeID string) *Runner {
	return &Runner{cancels: make(map[string]context.CancelFunc), emit: emit, nodeID: nodeID}
}

// Run executes a job asynchronously and streams its output.
func (r *Runner) Run(a protocol.JobAssign) {
	timeout := time.Duration(a.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = time.Hour
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	r.mu.Lock()
	r.cancels[a.JobID] = cancel
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		delete(r.cancels, a.JobID)
		r.mu.Unlock()
		cancel()
	}()

	cmd := exec.CommandContext(ctx, a.Command, a.Args...)
	if a.WorkDir != "" {
		cmd.Dir = a.WorkDir
	}
	for k, v := range a.Env {
		cmd.Env = append(cmd.Environ(), k+"="+v)
	}
	setSysProcAttr(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.fail(a.JobID, err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		r.fail(a.JobID, err)
		return
	}
	if err := cmd.Start(); err != nil {
		r.fail(a.JobID, err)
		return
	}

	var wg sync.WaitGroup
	var seq int64
	var seqMu sync.Mutex
	stream := func(name string, rc interface{ Read([]byte) (int, error) }) {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := rc.Read(buf)
			if n > 0 {
				seqMu.Lock()
				seq++
				s := seq
				seqMu.Unlock()
				r.emitOut(a.JobID, name, string(buf[:n]), s)
			}
			if err != nil {
				return
			}
		}
	}
	wg.Add(2)
	go stream("stdout", stdout)
	go stream("stderr", stderr)
	wg.Wait()

	waitErr := cmd.Wait()
	status := protocol.JobSuccess
	exitCode := 0
	if ctx.Err() == context.DeadlineExceeded {
		status = protocol.JobTimeout
		exitCode = -1
	} else if ctx.Err() == context.Canceled {
		status = protocol.JobCancelled
		exitCode = -1
	} else if waitErr != nil {
		status = protocol.JobFailed
		exitCode = exitCodeOf(waitErr)
	}
	r.result(a.JobID, status, exitCode, waitErr)
}

// Cancel terminates a running job if present.
func (r *Runner) Cancel(jobID string) {
	r.mu.Lock()
	cancel, ok := r.cancels[jobID]
	r.mu.Unlock()
	if ok {
		cancel()
	}
}

func (r *Runner) emitOut(jobID, streamName, data string, seq int64) {
	env, err := protocol.NewEnvelope(protocol.TypeJobOutput, fmt.Sprintf("%s-%d", jobID, seq),
		time.Now().Unix(), r.nodeID,
		protocol.JobOutput{JobID: jobID, Stream: streamName, Data: data, Seq: seq})
	if err == nil {
		r.emit.Send(env)
	}
}

func (r *Runner) result(jobID, status string, code int, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	env, e := protocol.NewEnvelope(protocol.TypeJobResult, "res-"+jobID, time.Now().Unix(), r.nodeID,
		protocol.JobResult{JobID: jobID, Status: status, ExitCode: code, Error: msg})
	if e == nil {
		r.emit.Send(env)
	}
}

func (r *Runner) fail(jobID string, err error) {
	r.result(jobID, protocol.JobFailed, -1, err)
}

func exitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
			return ws.ExitStatus()
		}
		return ee.ExitCode()
	}
	return -1
}
