package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
		"github.com/managerhub/managerhub/controller/internal/auth"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/shared/protocol"
)

// handleUIWS upgrades a browser connection for terminal + live updates.
// Protocol (JSON text frames):
//   → {"action":"terminal_open","node_id":"...","shell":"","cols":80,"rows":24}
//   → {"action":"terminal_input","session_id":"...","data":"<base64>"}
//   → {"action":"terminal_resize","session_id":"...","cols":120,"rows":40}
//   → {"action":"terminal_close","session_id":"..."}
//   → {"action":"job_create","name":"...","command":"...","node_id":"..."}
//   → {"action":"job_cancel","job_id":"..."}
//   ← {"event":"terminal_output","session_id":"...","data":"<base64>"}
//   ← {"event":"terminal_closed","session_id":"...","reason":"..."}
//   ← {"event":"job_output","job_id":"...","stream":"stdout","data":"..."}
//   ← {"event":"job_result","job_id":"...","status":"...","exit_code":0}
//   ← {"event":"node_status","node_id":"...","online":true}
func (s *Server) handleUIWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}
	claims, err := auth.ParseToken(s.Cfg.JWTSecret, token)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	user := claims.Subject
	s.Log.Info("ui ws connected", "user", user)

	// Register a broadcaster for this UI session
	bc := newBroadcaster(ws)
	s.addUIClient(bc)
	defer s.removeUIClient(bc)

	ws.SetReadLimit(64 * 1024)
	ws.SetReadDeadline(time.Now().Add(300 * time.Second))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(300 * time.Second))
		return nil
	})

	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			return
		}
		ws.SetReadDeadline(time.Now().Add(300 * time.Second))
		var msg uiMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		s.handleUIAction(r, bc, user, msg)
	}
}

type uiMessage struct {
	Action    string `json:"action"`
	NodeID    string `json:"node_id"`
	SessionID string `json:"session_id"`
	Shell     string `json:"shell"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
	Data      string `json:"data"`
	JobID     string `json:"job_id"`
	Name      string `json:"name"`
	Command   string `json:"command"`
	Args      []string `json:"args"`
}

func (s *Server) handleUIAction(r *http.Request, bc *broadcaster, user string, msg uiMessage) {
	ctx := r.Context()
	switch msg.Action {
	case "terminal_open":
		sessID := uuid.NewString()
		nodeID := msg.NodeID
		env, _ := protocol.NewEnvelope(protocol.TypeTermOpen, "term-open-"+sessID, time.Now().Unix(), nodeID,
			protocol.TermOpen{SessionID: sessID, Shell: msg.Shell, Cols: msg.Cols, Rows: msg.Rows})
		if err := s.Hub.Send(nodeID, env); err != nil {
			bc.send("error", map[string]string{"error": err.Error()})
			return
		}
		bc.trackTerminal(sessID, nodeID)
		bc.send("terminal_ready", map[string]string{"session_id": sessID})
		_ = s.Store.AppendAudit(ctx, user, "terminal.open", nodeID, "")

	case "terminal_input":
		bc.mu.Lock()
		nodeID := bc.terminals[msg.SessionID]
		bc.mu.Unlock()
		env, _ := protocol.NewEnvelope(protocol.TypeTermInput, "term-in-"+uuid.NewString(), time.Now().Unix(), nodeID,
			protocol.TermInput{SessionID: msg.SessionID, Data: msg.Data})
		_ = s.Hub.Send(nodeID, env)

	case "terminal_resize":
		bc.mu.Lock()
		nodeID := bc.terminals[msg.SessionID]
		bc.mu.Unlock()
		env, _ := protocol.NewEnvelope(protocol.TypeTermResize, "term-rz-"+uuid.NewString(), time.Now().Unix(), nodeID,
			protocol.TermResize{SessionID: msg.SessionID, Cols: msg.Cols, Rows: msg.Rows})
		_ = s.Hub.Send(nodeID, env)

	case "terminal_close":
		bc.mu.Lock()
		nodeID := bc.terminals[msg.SessionID]
		delete(bc.terminals, msg.SessionID)
		bc.mu.Unlock()
		env, _ := protocol.NewEnvelope(protocol.TypeTermClose, "term-cls-"+uuid.NewString(), time.Now().Unix(), nodeID,
			protocol.TermClose{SessionID: msg.SessionID})
		_ = s.Hub.Send(nodeID, env)
		_ = s.Store.AppendAudit(ctx, user, "terminal.close", nodeID, "")

	case "job_create":
		j, err := s.CreateAndDispatch(newJobFromUI(msg))
		if err != nil {
			bc.send("error", map[string]string{"error": err.Error()})
			return
		}
		bc.watchJob(j.ID)
		bc.send("job_created", map[string]string{"job_id": j.ID})
		_ = s.Store.AppendAudit(ctx, user, "job.create", j.ID, j.Name)

	case "job_cancel":
		env, _ := protocol.NewEnvelope(protocol.TypeJobCancel, "cancel-"+msg.JobID, time.Now().Unix(), "",
			protocol.JobCancel{JobID: msg.JobID})
		j, err := s.Store.GetJob(ctx, msg.JobID)
		if err == nil && j.NodeID != nil {
			_ = s.Hub.Send(*j.NodeID, env)
		}
		code := -1
		_ = s.Store.UpdateJobStatus(ctx, msg.JobID, protocol.JobCancelled, &code, "cancelled by user")

	case "job_watch":
		bc.watchJob(msg.JobID)

	case "services_list":
		nodeID := msg.NodeID
		env, _ := protocol.NewEnvelope(protocol.TypeSvcList, "svc-list-"+uuid.NewString(), time.Now().Unix(), nodeID,
			protocol.SvcListResult{})
		_ = s.Hub.Send(nodeID, env)

	case "services_action":
		nodeID := msg.NodeID
		env, _ := protocol.NewEnvelope(protocol.TypeSvcAction, "svc-act-"+uuid.NewString(), time.Now().Unix(), nodeID,
			protocol.SvcAction{ReqID: msg.SessionID, Name: msg.Data, Action: msg.Name})
		_ = s.Hub.Send(nodeID, env)

	case "runners_list":
		env, _ := protocol.NewEnvelope(protocol.TypeRunnerList, "run-list-"+uuid.NewString(), time.Now().Unix(), msg.NodeID,
			protocol.RunnerListResult{})
		_ = s.Hub.Send(msg.NodeID, env)

	case "runners_action":
		env, _ := protocol.NewEnvelope(protocol.TypeRunnerAction, "run-act-"+uuid.NewString(), time.Now().Unix(), msg.NodeID,
			protocol.RunnerAction{ReqID: msg.SessionID, Name: msg.Data, Action: msg.Name})
		_ = s.Hub.Send(msg.NodeID, env)
	}
}

func newJobFromUI(msg uiMessage) store.Job {
	if msg.Args == nil {
		msg.Args = []string{}
	}
	var nodeID *string
	if msg.NodeID != "" {
		nodeID = &msg.NodeID
	}
	return store.Job{
		Name: msg.Name, Type: "shell", Command: msg.Command,
		Args: msg.Args, TimeoutSec: 3600, NodeID: nodeID,
	}
}
