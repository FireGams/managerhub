package main

import (
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/managerhub/managerhub/agent/internal/runner"
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

	case protocol.TypeDockerList:
		a.handleDockerList(env)

	case protocol.TypeDockerAction:
		a.handleDockerAction(env)

	case protocol.TypeRunnerInstall:
		a.handleRunnerInstall(env)

	case protocol.TypeUninstall:
		a.log.Info("uninstall requested — stopping")
		a.term.CloseAll()
		os.Exit(0)

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

func (a *agent) handleDockerList(env protocol.Envelope) {
	avail := a.docker.Available()
	res := protocol.DockerListResult{ReqID: env.ID, Available: avail}
	if avail {
		list, err := a.docker.List()
		if err != nil {
			res.Error = err.Error()
		} else {
			for _, c := range list {
				res.Containers = append(res.Containers, map[string]string{
					"id": c.ID, "name": c.Name, "image": c.Image,
					"status": c.Status, "state": c.State,
				})
			}
		}
	}
	out, _ := protocol.NewEnvelope(protocol.TypeDockerListResult, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, res)
	a.cli.Send(out)
}

func (a *agent) handleDockerAction(env protocol.Envelope) {
	var req protocol.DockerAction
	if err := env.Decode(&req); err != nil {
		return
	}
	var logs string
	var err error
	if req.Action == "logs" {
		logs, err = a.docker.Logs(req.Name, 100)
	} else {
		err = a.docker.Action(req.Name, req.Action)
	}
	res := protocol.DockerActionResult{ReqID: env.ID, Name: req.Name, Action: req.Action,
		OK: err == nil, Logs: logs}
	if err != nil {
		res.Error = err.Error()
	}
	out, _ := protocol.NewEnvelope(protocol.TypeDockerActionResult, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, res)
	a.cli.Send(out)
}

func (a *agent) handleRunnerInstall(env protocol.Envelope) {
	var req protocol.RunnerInstall
	if err := env.Decode(&req); err != nil {
		return
	}
	output, err := runner.Install(req.RepoURL, req.Token, req.Name, req.Labels, req.WorkDir)
	res := protocol.RunnerInstallResult{ReqID: env.ID, OK: err == nil, Output: output}
	if err != nil {
		res.Error = err.Error()
	}
	out, _ := protocol.NewEnvelope(protocol.TypeRunnerInstallResult, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, res)
	a.cli.Send(out)
}

func (a *agent) sendError(refID, code string, err error) {
	out, _ := protocol.NewEnvelope(protocol.TypeError, uuid.NewString(), time.Now().Unix(),
		a.state.NodeID, protocol.ErrorPayload{Code: code, Message: err.Error(), RefID: refID})
	a.cli.Send(out)
}
