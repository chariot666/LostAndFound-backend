package handler

import "testing"

func TestValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{name: "letters and numbers", username: "zhang123", want: true},
		{name: "Chinese characters", username: "张三同学", want: true},
		{name: "mixed characters", username: "张三abc123", want: true},
		{name: "too short", username: "张三", want: false},
		{name: "too long", username: "abcdefghijk", want: false},
		{name: "contains underscore", username: "zhang_san", want: false},
		{name: "contains whitespace", username: "张三 同学", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validUsername(tt.username); got != tt.want {
				t.Fatalf("validUsername(%q) = %v, want %v", tt.username, got, tt.want)
			}
		})
	}
}

func TestValidPassword(t *testing.T) {
	tests := []struct {
		password string
		want     bool
	}{
		{password: "123456", want: true},
		{password: "12345", want: false},
		{password: "12345678901234567890", want: true},
		{password: "123456789012345678901", want: false},
	}

	for _, tt := range tests {
		if got := validPassword(tt.password); got != tt.want {
			t.Fatalf("validPassword(%q) = %v, want %v", tt.password, got, tt.want)
		}
	}
}
