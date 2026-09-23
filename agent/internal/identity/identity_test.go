package identity

import "testing"

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/s.json"
	if _, err := Load(path); err != ErrNotEnrolled {
		t.Fatalf("want ErrNotEnrolled, got %v", err)
	}
	want := State{NodeID: "n1", NodeToken: "t1"}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
}
