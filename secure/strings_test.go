package secure

import "testing"

func TestMaskString(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{"empty", "", ""},
		{"len1", "a", "*"},
		{"len2", "ab", "*b"},
		{"len3", "abc", "a*c"},
		{"len4", "abcd", "a**d"},
		{"len5", "abcde", "a**de"},
		{"len6", "abcdef", "ab**ef"},
		{"len7", "abcdefg", "ab***fg"},
		{"len8", "abcdefgh", "ab***fgh"},
		{"len9", "abcdefghi", "abc***ghi"},
		{"len10", "abcdefghij", "abc****hij"},
		{"len11", "abcdefghijk", "abc****hijk"},
		{"chinese", "你好，世界！", "你好**界！"},
		{"japanese", "こんにちは世界", "こん***世界"},
		{"korean", "안녕하세요세계", "안녕***세계"},
		{"emoji", "😀😃😄😁😆😅😂🤣", "😀😃***😅😂🤣"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskString(tt.in); got != tt.expected {
				t.Errorf("MaskString(%q) = %q, want %q", tt.in, got, tt.expected)
			}
		})
	}
}
