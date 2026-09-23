package protocol

import "errors"

// Message types: agent -> controller.
const (
	TypeHello              = "hello"
	TypeHeartbeat          = "heartbeat"
	TypeMetrics            = "metrics"
	TypePong               = "pong"
	TypeJobOutput          = "job_output"
	TypeJobResult          = "job_result"
	TypeTermOutput         = "term_output"
	TypeTermClosed         = "term_closed"
	TypeSvcListResult      = "svc_list_result"
	TypeSvcActionResult    = "svc_action_result"
	TypeRunnerListResult   = "runner_list_result"
	TypeRunnerActionResult = "runner_action_result"
	TypeDockerListResult   = "docker_list_result"
	TypeDockerActionResult = "docker_action_result"
	TypeError              = "error"
)

// Message types: controller -> agent.
const (
	TypeHelloAck     = "hello_ack"
	TypePing         = "ping"
	TypeJobAssign    = "job_assign"
	TypeJobCancel    = "job_cancel"
	TypeTermOpen     = "term_open"
	TypeTermInput    = "term_input"
	TypeTermResize   = "term_resize"
	TypeTermClose    = "term_close"
	TypeSvcList      = "svc_list"
	TypeSvcAction    = "svc_action"
	TypeRunnerList   = "runner_list"
	TypeRunnerAction = "runner_action"
	TypeDockerList   = "docker_list"
	TypeDockerAction = "docker_action"
)

// Job lifecycle states.
const (
	JobQueued    = "queued"
	JobAssigned  = "assigned"
	JobRunning   = "running"
	JobSuccess   = "success"
	JobFailed    = "failed"
	JobCancelled = "cancelled"
	JobTimeout   = "timeout"
)

// Node connection states.
const (
	NodeOnline  = "online"
	NodeOffline = "offline"
)

// Roles for RBAC.
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

// Sentinel errors of the protocol layer.
var (
	ErrUnsupportedVersion = errors.New("protocol: unsupported version")
	ErrUnknownType        = errors.New("protocol: unknown message type")
)
