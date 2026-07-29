package utils

import "testing"

func TestGenerateAlias(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		separator  string
		lowercase  bool
		partLength int
		want       string
	}{
		{
			name:       "simple path - first letter",
			path:       "/home/user/my-cool-project",
			separator:  "",
			lowercase:  true,
			partLength: 1,
			want:       "mcp",
		},
		{
			name:       "simple path - two letters",
			path:       "/home/user/my-cool-project",
			separator:  "",
			lowercase:  true,
			partLength: 2,
			want:       "mycopr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateAlias(tt.path, tt.separator, tt.lowercase, tt.partLength)
			if got != tt.want {
				t.Errorf("GenerateAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscapeAliasValue(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want string
	}{
		{
			name: "no quotes",
			val:  "echo hi",
			want: "echo hi",
		},
		{
			name: "with single quotes",
			val:  "echo 'hi'",
			want: "echo 'hi'",
		},
		{
			name: "with double quotes",
			val:  `echo "hi"`,
			want: "echo 'hi'",
		},
		{
			name: "with mixed quotes",
			val:  `echo "hello" and 'world'`,
			want: "echo 'hello' and 'world'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeAliasValue(tt.val)
			if got != tt.want {
				t.Errorf("EscapeAliasValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  []string
	}{
		{
			name:  "simple sentence",
			input: "hello world nice to meet you",
			width: 11,
			want:  []string{"hello world", "nice to", "meet you"},
		},
		{
			name:  "long word breaking",
			input: "thisisaverylongwordthatneedstobebroken",
			width: 10,
			want:  []string{"thisisaver", "ylongwordt", "hatneedsto", "bebroken"},
		},
		{
			name:  "empty lines preserved",
			input: "hello\n\nworld",
			width: 10,
			want:  []string{"hello", "", "world"},
		},
		{
			name:  "path wrapping",
			input: "/home/user/projects/clis/project-hub",
			width: 15,
			want:  []string{"/home/user/proj", "ects/clis/proje", "ct-hub"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapText(tt.input, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("WrapText() returned %d lines, want %d: %q vs %q", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
