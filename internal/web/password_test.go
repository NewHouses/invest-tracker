package web

import "testing"

func TestHashPasswordVerify(t *testing.T) {
	encoded, err := hashPassword("contrasinal-seguro", 1000)
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	ok, err := verifyPassword("contrasinal-seguro", encoded)
	if err != nil || !ok {
		t.Fatalf("verifyPassword correcto: ok=%v err=%v", ok, err)
	}
	ok, err = verifyPassword("contrasinal-errado", encoded)
	if err != nil || ok {
		t.Fatalf("verifyPassword errado: ok=%v err=%v", ok, err)
	}
}

func TestVerifyPasswordMalformed(t *testing.T) {
	cases := []string{
		"",
		"argon2$1$a$b",
		"pbkdf2-sha256$abc$a$b",
		"pbkdf2-sha256$1000$a$b",
		"pbkdf2-sha256$1000$AAAAAAAAAAAAAAAAAAAAAA==$b",
	}
	for _, tc := range cases {
		if ok, err := verifyPassword("contrasinal", tc); err == nil || ok {
			t.Fatalf("verifyPassword(%q) = ok=%v err=%v, esperabamos erro", tc, ok, err)
		}
	}
}

// O contrasinal non ten requisitos de lonxitude nin de complexidade: só non
// pode estar baleiro (e hai un límite técnico de 1024 bytes).
func TestValidatePasswordField(t *testing.T) {
	for _, pw := range []string{"1", "ab", "curto", "contrasinal longo con espazos", "ñ"} {
		if fields := validatePasswordField("password", pw); len(fields) != 0 {
			t.Errorf("validatePasswordField(%q) = %v, esperabamos válido", pw, fields)
		}
	}
	if fields := validatePasswordField("password", ""); fields["password"] == "" {
		t.Errorf("o contrasinal baleiro debe rexeitarse, got %v", fields)
	}
	long := string(make([]byte, passwordMaxBytes+1))
	if fields := validatePasswordField("password", long); fields["password"] == "" {
		t.Errorf("un contrasinal de máis de %d bytes debe rexeitarse", passwordMaxBytes)
	}
}
