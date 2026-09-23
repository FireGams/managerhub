// Package protocol contains the versioned wire format shared by the
// controller and the agents. It must stay free of I/O and OS-specific code
// so both sides (and future third-party agents) can depend on it safely.
package protocol
