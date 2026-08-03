package platform

import "testing"

func TestRelevantProcess(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{"default code", "/Applications/Visual Studio Code.app/Contents/MacOS/Electron", true},
		{"selected runtime", "/Applications/Code --user-data-dir /tmp/dist/old/.data", true},
		{"other runtime", "/Applications/Code --user-data-dir /tmp/dist/other/.data", false},
		{"helper", "/Applications/Code Helper --user-data-dir /tmp/dist/old/.data", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := relevantProcess("code", tt.args, []string{"/tmp/dist/old/.data", "/tmp/dist/new/.data"}); got != tt.want {
				t.Fatalf("relevantProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}
