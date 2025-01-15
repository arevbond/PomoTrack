package config

//func TestInit(t *testing.T) {
//	exe := filepath.Join(t.TempDir(), "pomotrack")
//	build := exec.Command("go", "build", "-o", exe)
//	if err := build.Run(); err != nil {
//		t.Fatalf("failed to build: %v", err)
//	}
//
//	tests := []struct {
//		name     string
//		args     []string
//		expected *Config
//	}{
//		{
//			name: "default config without cmd arguments",
//			args: make([]string, 0),
//			expected: &Config{
//				Timer: TimerConfig{
//					FocusDuration: defaultFocusDuration,
//					BreakDuration: defaultBreakDuration,
//					HiddenFocusTime: true,
//				},
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//		})
//	}
//}
//
//func TestParseFlags(t *testing.T) {
//	exe := filepath.Join(t.TempDir(), "pomotrack")
//	build := exec.Command("go", "build", "-o", exe)
//	if err := build.Run(); err != nil {
//		t.Fatalf("failed to build: %v", err)
//	}
//
//	tests := []struct {
//		name     string
//		args     []string
//		expected flags
//	}{
//		{
//			name: "default flags without cmd arguments",
//			args: []string{"cmd"},
//			expected: flags{
//				FocusDuration: 0,
//				BreakDuration: 0,
//				HiddenFocusTime: true,
//			},
//		},
//		{
//			name: "correct applying flags",
//			args: []string{"cmd", "--focus-duration=25m", "--break-duration=5m", "--show-focus=true"},
//			expected: flags{
//				FocusDuration: 25 * time.Minute,
//				BreakDuration: 5 * time.Minute,
//				HiddenFocusTime: false,
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			os.Args = tt.args
//			result := parseFlags()
//			require.Equal(t, tt.expected, result)
//		})
//	}
//}
