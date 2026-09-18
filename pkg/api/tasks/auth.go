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
	"github.com/pyromeowww/task-scheduler/pkg/settings"
)

// Config хранит настройки аутентификации, прочитанные один раз на старте.
type Config struct {
	PasswordHash string        // SHA-256 от пароля из окружения.
	JWTSecret    []byte        // ключ для подписи и проверки HS256-токенов.
	TokenTTL     time.Duration // время жизни выдаваемых токенов.
}

func LoadConfig() Config {
	password := os.Getenv(settings.EnvTodoPassword)
	cfg := Config{
		JWTSecret: []byte(os.Getenv(settings.EnvSaltJWT)),
		TokenTTL:  8 * time.Hour,
	}
	if password != "" {
		cfg.PasswordHash = hashPassword(password)
	}
	return cfg
}

// SigninHandler обрабатывает POST. Проверяет пароль и возвращает JWT
func SigninHandler(cfg Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			writeError(res, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Cтруктура под тело запроса.
		var creds struct {
			Password string `json:"password"`
		}

		// Читаем JSON из тела.
		if err := json.NewDecoder(req.Body).Decode(&creds); err != nil {
			writeError(res, http.StatusBadRequest, "JSON deserialization error: "+err.Error())
			return
		}

		if cfg.PasswordHash == "" {
			writeError(res, http.StatusUnauthorized, "Аутентификация не настроена")
			return
		}
		// Сравниваем присланный пароль с ожидаемым.
		if subtle.ConstantTimeCompare([]byte(hashPassword(creds.Password)), []byte(cfg.PasswordHash)) != 1 {
			writeError(res, http.StatusUnauthorized, "Неверный пароль")
			return
		}
		// Создаём токен с алгоритмом HS256 и подписываем его секретом.
		signed, err := createJWT(cfg.PasswordHash, cfg.JWTSecret, cfg.TokenTTL)
		if err != nil {
			writeError(res, http.StatusInternalServerError, "Ошибка создания токена")
			return
		}
		writeJSON(res, http.StatusOK, map[string]string{"token": signed})
	}
}

// Auth возвращает middleware, проверяющий JWT из куки
func Auth(cfg Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if cfg.PasswordHash == "" {
			next.ServeHTTP(res, req)
			return
		}
		cookie, err := req.Cookie("token")
		if err != nil {
			http.Error(res, settings.ErrAuthRequired, http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(cookie.Value,
			func(t *jwt.Token) (interface{}, error) { return cfg.JWTSecret, nil },
			jwt.WithValidMethods([]string{"HS256"}),
		)
		if err != nil || !token.Valid {
			http.Error(res, settings.ErrAuthRequired, http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(res, settings.ErrAuthRequired, http.StatusUnauthorized)
			return
		}
		hashFromToken, ok := claims["hash"].(string)
		if !ok {
			http.Error(res, settings.ErrAuthRequired, http.StatusUnauthorized)
			return
		}
		if subtle.ConstantTimeCompare([]byte(hashFromToken), []byte(cfg.PasswordHash)) != 1 {
			http.Error(res, settings.ErrAuthRequired, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(res, req)
	})
}

func hashPassword(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func createJWT(passwordHash string, secret []byte, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"hash": passwordHash,
		"exp":  time.Now().Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
