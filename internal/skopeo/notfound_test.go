package skopeo

import "testing"

func TestIsDaemonNotFound(t *testing.T) {
	cases := []struct {
		msg  string
		want bool
	}{
		{"Error response from daemon: No such image: foo:bar", true},
		{"FATA[0000] ... no such image: foo", true},
		{"image not found", true},
		{"does not exist", true},
		{"Cannot connect to the Docker daemon at unix:///var/run/docker.sock", false},
		{"permission denied", false},
		{"", false},
	}
	for _, tt := range cases {
		if got := isDaemonNotFound(tt.msg); got != tt.want {
			t.Fatalf("isDaemonNotFound(%q)=%v want %v", tt.msg, got, tt.want)
		}
	}
}
