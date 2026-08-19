package auth

import "testing"

func TestNormalizeUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "lowercase", input: "Alice_01", want: "alice_01"},
		{name: "valid dot", input: "ab.cd", want: "ab.cd"},
		{name: "leading dot", input: ".abcd", wantErr: true},
		{name: "trailing dot", input: "abcd.", wantErr: true},
		{name: "double dot", input: "ab..cd", wantErr: true},
		{name: "invalid symbol", input: "ab-cd", wantErr: true},
		{name: "too short", input: "abc", wantErr: true},
		{name: "reserved", input: "admin", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeUsername(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("got %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("expected short password to fail")
	}
	if err := ValidatePassword("correct horse battery staple"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
