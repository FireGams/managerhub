package protocol

import "testing"

func TestEnvelopeRoundtrip(t *testing.T) {
	env, err := NewEnvelope(TypeHeartbeat, "m1", 42, "n1", Heartbeat{UptimeSec: 7})
	if err != nil { t.Fatal(err) }
	raw, err := env.Encode()
	if err != nil { t.Fatal(err) }
	got, err := Parse(raw)
	if err != nil { t.Fatal(err) }
	if got.Type != TypeHeartbeat || got.NodeID != "n1" || got.Version != CurrentVersion {
		t.Fatalf("bad envelope: %+v", got)
	}
	var hb Heartbeat
	if err := got.Decode(&hb); err != nil { t.Fatal(err) }
	if hb.UptimeSec != 7 { t.Fatalf("uptime=%d", hb.UptimeSec) }
}

func TestParseRejectsBadVersion(t *testing.T) {
	if _, err := Parse([]byte(`{"v":99,"type":"x"}`)); err != ErrUnsupportedVersion {
		t.Fatalf("want ErrUnsupportedVersion, got %v", err)
	}
}
