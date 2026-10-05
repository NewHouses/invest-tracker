package web

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"crypto/pbkdf2"
)

const (
	passwordSettingKey = "password_hash"
	defaultPBKDF2Iter  = 600000
	passwordSaltBytes  = 16
	passwordKeyBytes   = 32
	passwordMaxBytes   = 1024
)

func hashPassword(pw string, iter int) (string, error) {
	if iter <= 0 {
		return "", fmt.Errorf("as iteracións deben ser positivas")
	}
	salt := make([]byte, passwordSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, pw, salt, iter, passwordKeyBytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", iter, base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString(key)), nil
}

func verifyPassword(pw, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false, fmt.Errorf("formato de contrasinal non válido")
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false, fmt.Errorf("iteracións de contrasinal non válidas")
	}
	salt, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) != passwordSaltBytes {
		return false, fmt.Errorf("sal de contrasinal non válida")
	}
	want, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil || len(want) != passwordKeyBytes {
		return false, fmt.Errorf("hash de contrasinal non válido")
	}
	got, err := pbkdf2.Key(sha256.New, pw, salt, iter, len(want))
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// validatePasswordField non impón requisitos de lonxitude nin de
// complexidade: só rexeita o contrasinal baleiro (deixaría a aplicación aberta
// a toda a rede local) e os de máis de 1024 bytes (límite técnico).
func validatePasswordField(field, pw string) map[string]string {
	fields := map[string]string{}
	switch {
	case pw == "":
		fields[field] = "o contrasinal non pode estar baleiro"
	case len(pw) > passwordMaxBytes:
		fields[field] = "o contrasinal non pode superar 1024 bytes"
	}
	return fields
}
