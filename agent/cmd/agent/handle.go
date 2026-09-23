package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/managerhub/managerhub/shared/protocol"
)

// handle processes one inbound envelope from the controller.
func (a *agent) handle(env protocol.Envelope) {
	switch env.Type {
	case protocol.TypeHelloAck:
		a.log.Debug("hello acknowledged")

	case protocol.TypePing:
		var p protocol.PingPong
		_ = env.Decode(&p)
		out, _ := protocol.NewEnvelope(protocol.TypePong, uuid.NewString(), time.Now().Unix(),
			a.state.NodeID, p)
		a.cli.Send(out)

	case protocol.TypeJobAssign:
		var j protocol.JobAssign
		if err := env.Decode(&j); err != nil {
			a.log.Warn("bad job assign", "err", err)
			return
		}
		a.log.Info("job received", "job_id", j.JobID, "name", j.Name)
		go a.jobs.Run(j)

	case protocol.TypeJobCancel:
		var c protocol.JobCancel
		if err := env.Decode(&c); err == nil {
			a.jobs.Cancel(c.JobID)
		}

	case protocol.TypeTermOpen:
		var t protocol.TermOpen
		if err := env.Decode(&t); err == nil {
			if err := a.term.Open(t); err != nil {
				a.sendError(env.ID, "term_open", err)
			}
		}

	case protocol.TypeTermInput:
		var t protocol.TermInput
		if err := env.Decode(&t); err == nil {
			_ = a.term.Input(t)
		}

	case protocol.TypeTermResize:
		var t protocol.TermResize
		if err := env.Decode(&t); err == nil {
			_ = a.term.Resize(t)
		}

	case protocol.TypeTermClose:
		var t protocol.TermClose
		if err := env.Decode(&t); err == nil {
			a.term.Close(t.SessionID, "controller close")
		}

	case protocol.TypeSvcList:
		a.handleSvcList(env)

	case protocol.TypeSvcAction:
		a.handleSvcAction(env)

	case protocol.TypeRunnerList:
		a.handleRunnerList(env)

	case protocol.TypeRunnerAction:
		a.handleRunnerAction(env)

	default:
		a.log.Debug("unhandled message", "type", env.Type)
	}
}

func (a *agent) handleSvcList(env protocol.Envelope) {
	list, err := a.svc.List()
	res := protocol.SvcListResult{ReqID: env.ID, Services: list}
	if err != nil {
		res.Error = err.Error()
	}
	out, _ := protocol.NewEnvelope(protocol.TypeSvcListResult, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, res)
	a.cli.Send(out)
}

func (a *agent) handleSvcAction(env protocol.Envelope) {
	var req protocol.SvcAction
	if err := env.Decode(&req); err != nil {
		return
	}
	err := a.svc.Action(req.Name, req.Action)
	res := protocol.SvcActionResult{ReqID: env.ID, Name: req.Name, Action: req.Action, OK: err == nil}
	if err != nil {
		res.Error = err.Error()
	}
	out, _ := protocol.NewEnvelope(protocol.TypeSvcActionResult, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, res)
	a.cli.Send(out)
}

func (a *agent) handleRunnerList(env protocol.Envelope) {
	list, err := a.det.List()
	res := protocol.RunnerListResult{ReqID: env.ID, Runners: list}
	if err != nil {
		res.Error = err.Error()
	}
	out, _ := protocol.NewEnvelope(protocol.TypeRunnerListResult, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, res)
	a.cli.Send(out)
}

func (a *agent) handleRunnerAction(env protocol.Envelope) {
	var req protocol.RunnerAction
	if err := env.Decode(&req); err != nil {
		return
	}
	logs, err := a.det.Action(req.Name, req.Action)
	res := protocol.RunnerActionResult{ReqID: env.ID, Name: req.Name, Action: req.Action,
		OK: err == nil, Logs: logs}
	if err != nil {
		res.Error = err.Error()
	}
	out, _ := protocol.NewEnvelope(protocol.TypeRunnerActionResult, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, res)
	a.cli.Send(out)
}

func (a *agent) sendError(refID, code string, err error) {
	out, _ := protocol.NewEnvelope(protocol.TypeError, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, protocol.ErrorPayload{Code: code, Message: err.Error(), RefID: refID})
	a.cli.Send(out)
	_ = fmt.Sprintf("")
}
