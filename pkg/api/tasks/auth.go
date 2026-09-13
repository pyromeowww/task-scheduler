package tasks

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SigninHandler обрабатывает POST.
func SigninHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Cтруктура под тело запроса.
	var creds struct {
		Password string `json:"password"`
	}

	defer req.Body.Close()

	// Читаем JSON из тела.
	if err := json.NewDecoder(req.Body).Decode(&creds); err != nil {
		writeError(res, http.StatusBadRequest, "JSON deserialization error: "+err.Error())
		return
	}

	// Берём пароль из окружения.
	expected := os.Getenv("TODO_PASSWORD")
	if expected == "" {
		writeError(res, http.StatusUnauthorized, "Аутентификация не настроена")
		return
	}
	// Сравниваем присланный пароль с ожидаемым.
	if subtle.ConstantTimeCompare([]byte(creds.Password), []byte(expected)) != 1 {
		writeError(res, http.StatusUnauthorized, "Неверный пароль")
		return
	}

	// Считаем SHA-256 от пароля и превращаем в hex-строку.
	sum := sha256.Sum256([]byte(creds.Password))
	hashHex := hex.EncodeToString(sum[:])
	claims := jwt.MapClaims{
		"hash": hashHex,
		"exp":  time.Now().Add(8 * time.Hour).Unix(),
	}

	// Создаём токен с алгоритмом HS256 и подписываем его секретом.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("SALT_JWT")))
	if err != nil {
		writeError(res, http.StatusInternalServerError, "Ошибка создания токена")
		return
	}
	writeJSON(res, http.StatusOK, map[string]string{"token": signed})
}
