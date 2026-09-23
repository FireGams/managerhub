package auth

import (
	"testing"
	"time"
)

func TestPasswordHash(t *testing.T) {
	h, err := HashPassword("s3cret-pw")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(h, "s3cret-pw") {
		t.Fatal("valid password must match")
	}
	if CheckPassword(h, "wrong") {
		t.Fatal("wrong password must not match")
	}
}

func TestJWT(t *testing.T) {
	tok, err := IssueToken("secret-key-123456", "alice", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseToken("secret-key-123456", tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.Subject != "alice" || c.Role != "admin" {
		t.Fatalf("claims=%+v", c)
	}
	if _, err := ParseToken("other-secret-key!", tok); err == nil {
		t.Fatal("wrong secret must fail")
	}
}

func TestRoleAtLeast(t *testing.T) {
	cases := []struct {
		role, min string
		want      bool
	}{
		{"admin", "operator", true},
		{"operator", "operator", true},
		{"viewer", "operator", false},
		{"admin", "admin", true},
		{"viewer", "admin", false},
	}
	for _, c := range cases {
		if got := RoleAtLeast(c.role, c.min); got != c.want {
			t.Fatalf("RoleAtLeast(%s,%s)=%v want %v", c.role, c.min, got, c.want)
		}
	}
}
