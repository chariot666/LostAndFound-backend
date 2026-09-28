package handler

import "testing"

func TestCanClaimItem(t *testing.T) {
	tests := []struct {
		itemType string
		want     bool
	}{
		{itemType: "found", want: true},
		{itemType: "lost", want: false},
		{itemType: "", want: false},
	}

	for _, tt := range tests {
		if got := canClaimItem(tt.itemType); got != tt.want {
			t.Errorf("canClaimItem(%q) = %v, want %v", tt.itemType, got, tt.want)
		}
	}
}
