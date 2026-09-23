package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/controller/internal/hub"
	"github.com/managerhub/managerhub/shared/protocol"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(*http.Request) bool { return true },
}

// handleAgentWS upgrades an authenticated agent connection.
func (s *Server) handleAgentWS(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	node, err := s.Store.GetNodeByToken(r.Context(), hashSecret(token))
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn := hub.NewConn(node.ID, ws)
	s.Hub.Register(conn)
	bg := context.Background()
	_ = s.Store.SetNodeStatus(bg, node.ID, protocol.NodeOnline)
	_ = s.Store.AppendNodeEvent(bg, node.ID, "connected", r.RemoteAddr)
	geo := lookupGeo(bg, readClientIP(r))
	_, _ = s.Store.Pool.Exec(bg, `UPDATE nodes SET city=$2, country=$3 WHERE id=$1`, node.ID, geo.City, geo.Country)
	s.Log.Info("agent connected", "node", node.Name, "id", node.ID)

	defer func() {
		s.Hub.Unregister(conn)
		_ = s.Store.SetNodeStatus(bg, node.ID, protocol.NodeOffline)
		_ = s.Store.AppendNodeEvent(bg, node.ID, "disconnected", "")
		s.Log.Info("agent disconnected", "node", node.Name, "id", node.ID)
	}()

	ack, _ := protocol.NewEnvelope(protocol.TypeHelloAck, "ack-"+node.ID, time.Now().Unix(), node.ID,
		protocol.HelloAck{NodeID: node.ID, HeartbeatInterval: 10, MetricsInterval: 15, ServerTime: time.Now().Unix()})
	_ = conn.Send(ack)

	ws.SetReadLimit(1 << 20)
	_ = ws.SetReadDeadline(time.Now().Add(90 * time.Second))
	ws.SetPongHandler(func(string) error {
		_ = ws.SetReadDeadline(time.Now().Add(90 * time.Second))
		return nil
	})

	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			return
		}
		_ = ws.SetReadDeadline(time.Now().Add(90 * time.Second))
		env, err := protocol.Parse(raw)
		if err != nil {
			s.Log.Warn("bad envelope", "err", err)
			continue
		}
		if conn.Duplicate(env.ID) {
			continue
		}
		s.dispatchAgentMessage(node.ID, env)
	}
}

func (s *Server) dispatchAgentMessage(nodeID string, env protocol.Envelope) {
	ctx := context.Background()
	switch env.Type {
	case protocol.TypeHello:
		var h protocol.Hello
		if err := env.Decode(&h); err == nil {
			_, _ = s.Store.UpsertNode(ctx, store.Node{
				ID: nodeID, Name: h.Name, Hostname: h.Hostname, OS: h.OS, Arch: h.Arch,
				IP: h.IP, AgentVersion: h.AgentVer, Tags: h.Tags, Capabilities: h.Capabilities,
			}, "")
		}
	case protocol.TypeHeartbeat:
		_ = s.Store.TouchNode(ctx, nodeID)
	case protocol.TypeMetrics:
		var m protocol.Metrics
		if err := env.Decode(&m); err == nil {
			gpus, _ := json.Marshal(m.GPUs)
			_ = s.Store.InsertMetrics(ctx, store.Metrics{
				NodeID: nodeID, TS: time.Now(), CPUPercent: m.CPUPercent, CPUCores: m.CPUCores,
				Load1: m.Load1, RAMUsed: m.RAMUsed, RAMTotal: m.RAMTotal, DiskUsed: m.DiskUsed,
				DiskTotal: m.DiskTotal, UptimeSec: m.UptimeSec, GPUs: string(gpus),
			})
		}
	case protocol.TypeJobOutput:
		var o protocol.JobOutput
		if err := env.Decode(&o); err == nil {
			_ = s.Store.AppendJobOutput(ctx, o.JobID, o.Stream, o.Data)
			s.broadcastJobEvent(o.JobID, "job_output", o)
		}
	case protocol.TypeJobResult:
		var res protocol.JobResult
		if err := env.Decode(&res); err == nil {
			code := res.ExitCode
			_ = s.Store.UpdateJobStatus(ctx, res.JobID, res.Status, &code, res.Error)
			s.broadcastJobEvent(res.JobID, "job_result", res)
		}
	case protocol.TypeTermOutput:
		var t protocol.TermOutput
		if err := env.Decode(&t); err == nil {
			s.forwardTerminal(t.SessionID, "terminal_output", t)
		}
	case protocol.TypeTermClosed:
		var t protocol.TermClosed
		if err := env.Decode(&t); err == nil {
			s.forwardTerminal(t.SessionID, "terminal_closed", t)
		}
	case protocol.TypeSvcListResult:
		var r protocol.SvcListResult
		if err := env.Decode(&r); err == nil {
			s.broadcastUI("services_list_result", r)
		}
	case protocol.TypeSvcActionResult:
		var r protocol.SvcActionResult
		if err := env.Decode(&r); err == nil {
			s.broadcastUI("services_action_result", r)
		}
	case protocol.TypeRunnerListResult:
		var r protocol.RunnerListResult
		if err := env.Decode(&r); err == nil {
			s.broadcastUI("runners_list_result", r)
		}
	case protocol.TypeRunnerActionResult:
		var r protocol.RunnerActionResult
		if err := env.Decode(&r); err == nil {
			s.broadcastUI("runners_action_result", r)
		}
	default:
		s.Log.Debug("unhandled agent message", "type", env.Type)
	}
}

func (s *Server) forwardTerminal(sessID, event string, payload any) {
	s.uiMu.RLock()
	clients := make([]*broadcaster, 0, len(s.uiClients))
	for bc := range s.uiClients {
		bc.mu.Lock()
		_, ok := bc.terminals[sessID]
		bc.mu.Unlock()
		if ok {
			clients = append(clients, bc)
		}
	}
	s.uiMu.RUnlock()
	for _, bc := range clients {
		bc.send(event, payload)
	}
}
