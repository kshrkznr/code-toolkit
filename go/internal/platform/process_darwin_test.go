package platform

import "testing"

func TestRelevantProcess(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		args     string
		want     bool
	}{
		{"default code", "code", "/Applications/Visual Studio Code.app/Contents/MacOS/Electron", true},
		{"selected runtime", "code", "/Applications/Code --user-data-dir /tmp/dist/old/.data", true},
		{"other runtime", "code", "/Applications/Code --user-data-dir /tmp/dist/other/.data", false},
		{"helper", "code", "/Applications/Code Helper --user-data-dir /tmp/dist/old/.data", false},
		{"default cursor", "cursor", "/Applications/Cursor.app/Contents/MacOS/Cursor", true},
		{"cursor helper", "cursor", "/Applications/Cursor.app/Contents/Frameworks/Cursor Helper.app/Contents/MacOS/Cursor Helper --type=gpu-process", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := relevantProcess(tt.platform, tt.args, []string{"/tmp/dist/old/.data", "/tmp/dist/new/.data"}); got != tt.want {
				t.Fatalf("relevantProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}
