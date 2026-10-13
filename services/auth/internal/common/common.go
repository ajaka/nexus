package common

import (
	"auth/internal/cache"
	"auth/internal/configs"
	"auth/internal/errs"
	"auth/internal/models"
	"auth/internal/store"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func BuildPayload(email, link, token string) ([]byte, error) {
	url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	query := url.Query()
	query.Set("token", token)
	url.RawQuery = query.Encode()

	p := models.KafkaPayload{
		Email: email,
		Url:   *url,
	}
	return json.Marshal(p)
}

func GetLogger(c *gin.Context) *slog.Logger {
	if l, ok := c.Get("logger"); ok {
		if logger, ok := l.(*slog.Logger); ok && logger != nil {
			return logger
		}
	}
	return slog.Default()
}
func GetFromContext[T any](c *gin.Context, key string) (T, bool) {
	var data T
	val, ok := c.Get(key)
	if !ok {
		return data, false
	}
	typedCasted, ok := val.(T)
	if !ok || iszero(val) {
		return data, false
	}
	return typedCasted, true
}

func iszero(d any) bool {
	if d == nil {
		return true
	}
	rv := reflect.ValueOf(d)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	}
	return reflect.DeepEqual(d, reflect.Zero(rv.Type()).Interface())
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, err
}

func HoursToDuration(hours float64) time.Duration {
	return time.Duration(hours * float64(time.Hour))
}

func GenerateJWT(secret string, duration float64, payload jwt.Claims) (string, time.Duration, error) {
	if duration <= 0 {
		return "", 0, fmt.Errorf("JWT duration must be greater than zero")
	}

	tokenDuration := HoursToDuration(duration)
	expiresAt := jwt.NewNumericDate(time.Now().Add(tokenDuration))
	switch claims := payload.(type) {
	case *models.LoneEmailPayload:
		claims.ExpiresAt = expiresAt
	default:
		return "", 0, fmt.Errorf("unsupported JWT claims type %T", payload)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return signedToken, tokenDuration, nil
}

func VerifyJWT(ctx context.Context, c *cache.Cache, tokenString, prefix, secret string, claims jwt.Claims) error {
	blacklisted, err := c.CheckBlackList(ctx, prefix, TokenDigest(tokenString))
	if err != nil {
		return err
	}
	if blacklisted {
		return errs.ERR_BLACKLISTED_TOKEN
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ERR_INVALID_METHOD
		}
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrTokenInvalidClaims
	}
	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		return err
	}
	if expiresAt == nil || expiresAt.Time.Before(time.Now()) {
		return jwt.ErrTokenExpired
	}
	return nil
}

func SetCookie(c *gin.Context, name, value string, maxAge int, secure bool) {
	c.SetCookie(name, value, maxAge, "/", "", secure, true)
}

func GenerateCleanUUID() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")
}

func HandleLoginActivity(c *gin.Context, payload *models.MinimalUserStruct, s *store.Store, session *models.Sessions, production bool) error {
	id := GenerateCleanUUID()
	session.SessionId = id
	exp, err := s.SetUserOnline(c.Request.Context(), payload, session)
	if err != nil {
		return err
	}
	expInInt := int(time.Until(exp).Seconds())
	SetCookie(c, "sessionId", id, expInInt, production)
	return nil
}

func TokenDigest(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	digest := h.Sum(nil)
	return hex.EncodeToString(digest[:])
}

func BlacklistToken(ctx context.Context, c *cache.Cache, token, prefix string, duration time.Duration) error {
	if !c.AddToBlacklist(ctx, prefix, TokenDigest(token), duration) {
		return errs.ERR_BLACKLIST_TOKEN_FAILURE
	}
	return nil
}

func ValidatePasswordLength(password string) bool {
	if len([]byte(password)) > 50 {
		return false
	}
	return true
}

func HandleLogoutActivity(c *gin.Context, s *store.Store, env *configs.Env, sessionId string, user *models.MinimalUserStruct) error {

	err := s.SetUserOffline(c.Request.Context(), sessionId, user.UserId)
	if err != nil {
		return err
	}
	c.SetSameSite(http.SameSiteLaxMode)
	SetCookie(c, "sessionId", "", -1, env.PRODUCTION)
	return nil
}
