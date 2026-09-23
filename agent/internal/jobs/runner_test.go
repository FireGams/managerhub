package jobs

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/managerhub/managerhub/shared/protocol"
)

type fakeEmit struct {
	mu   sync.Mutex
	envs []protocol.Envelope
}

func (f *fakeEmit) Send(env protocol.Envelope) {
	f.mu.Lock()
	f.envs = append(f.envs, env)
	f.mu.Unlock()
}

func (f *fakeEmit) find(typ string) []protocol.Envelope {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []protocol.Envelope
	for _, e := range f.envs {
		if e.Type == typ {
			out = append(out, e)
		}
	}
	return out
}

func waitResult(t *testing.T, f *fakeEmit) protocol.JobResult {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for job result")
		case <-time.After(20 * time.Millisecond):
			res := f.find(protocol.TypeJobResult)
			if len(res) > 0 {
				var r protocol.JobResult
				if err := res[0].Decode(&r); err != nil {
					t.Fatal(err)
				}
				return r
			}
		}
	}
}

func TestRunSuccess(t *testing.T) {
	f := &fakeEmit{}
	r := NewRunner(f, "n1")
	r.Run(protocol.JobAssign{
		JobID: "j1", Name: "echo", Type: "shell",
		Command: "/bin/echo", Args: []string{"hello"}, TimeoutSec: 5,
	})
	res := waitResult(t, f)
	if res.Status != protocol.JobSuccess || res.ExitCode != 0 {
		t.Fatalf("result=%+v", res)
	}
	outs := f.find(protocol.TypeJobOutput)
	if len(outs) == 0 {
		t.Fatal("expected stdout output")
	}
	var o protocol.JobOutput
	_ = outs[0].Decode(&o)
	if !strings.Contains(o.Data, "hello") {
		t.Fatalf("output=%q", o.Data)
	}
}

func TestRunFailure(t *testing.T) {
	f := &fakeEmit{}
	r := NewRunner(f, "n1")
	r.Run(protocol.JobAssign{
		JobID: "j2", Type: "shell", Command: "/bin/false", TimeoutSec: 5,
	})
	res := waitResult(t, f)
	if res.Status != protocol.JobFailed {
		t.Fatalf("result=%+v", res)
	}
}

func TestRunTimeout(t *testing.T) {
	f := &fakeEmit{}
	r := NewRunner(f, "n1")
	r.Run(protocol.JobAssign{
		JobID: "j3", Type: "shell", Command: "/bin/sleep", Args: []string{"10"}, TimeoutSec: 1,
	})
	res := waitResult(t, f)
	if res.Status != protocol.JobTimeout {
		t.Fatalf("result=%+v", res)
	}
}

func TestCancel(t *testing.T) {
	f := &fakeEmit{}
	r := NewRunner(f, "n1")
	done := make(chan struct{})
	go func() {
		r.Run(protocol.JobAssign{
			JobID: "j4", Type: "shell", Command: "/bin/sleep", Args: []string{"30"}, TimeoutSec: 60,
		})
		close(done)
	}()
	time.Sleep(200 * time.Millisecond)
	r.Cancel("j4")
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("cancel did not terminate job")
	}
	res := waitResult(t, f)
	if res.Status != protocol.JobCancelled {
		t.Fatalf("result=%+v", res)
	}
}
