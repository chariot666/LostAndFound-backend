package auth

import "testing"

func TestGenerateAndParseToken(t *testing.T) {
	token, err := GenerateToken("test-secret", 10001, "zhangsan", "user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ParseToken("test-secret", token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UID != 10001 || claims.Username != "zhangsan" || claims.Role != "user" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	token, err := GenerateToken("correct-secret", 10001, "zhangsan", "user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := ParseToken("wrong-secret", token); err == nil {
		t.Fatal("ParseToken() expected an error for wrong secret")
	}
}
