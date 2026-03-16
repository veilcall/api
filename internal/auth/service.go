package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	base58Alphabet   = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	recoveryCodeLen  = 12
	sessionTokenLen  = 32
	sessionTTL       = 24 * time.Hour
	maxLoginAttempts = 5
	attemptWindow    = 15 * time.Minute
)

type Service struct {
	repo   *Repository
	rdb    *redis.Client
	secret string
}

func NewService(repo *Repository, rdb *redis.Client, secret string) *Service {
	return &Service{repo: repo, rdb: rdb, secret: secret}
}

func (s *Service) Register(ctx context.Context) (userID, recoveryCode string, err error) {
	recoveryCode, err = generateBase58(recoveryCodeLen)
	if err != nil {
		return "", "", err
	}
	hash := s.hashRecovery(recoveryCode)
	userID, err = s.repo.CreateUser(ctx, hash)
	if err != nil {
		return "", "", err
	}
	return userID, recoveryCode, nil
}

func (s *Service) Login(ctx context.Context, recoveryCode string) (string, error) {
	hash := s.hashRecovery(recoveryCode)
	prefix := hash[:16]

	attemptKey := fmt.Sprintf("login_attempt:%s", prefix)
	count, err := s.rdb.Incr(ctx, attemptKey).Result()
	if err != nil {
		return "", fmt.Errorf("rate limit check: %w", err)
	}
	if count == 1 {
		s.rdb.Expire(ctx, attemptKey, attemptWindow)
	}
	if count > maxLoginAttempts {
		return "", fmt.Errorf("too many attempts, try again later")
	}

	userID, err := s.repo.GetUserByHash(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("invalid recovery code")
	}

	token, err := generateHex(sessionTokenLen)
	if err != nil {
		return "", err
	}

	sessKey := fmt.Sprintf("sess:%s", token)
	if err := s.rdb.Set(ctx, sessKey, userID, sessionTTL).Err(); err != nil {
		return "", fmt.Errorf("store session: %w", err)
	}

	s.rdb.Del(ctx, attemptKey)
	return token, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("sess:%s", token)).Err()
}

func (s *Service) hashRecovery(code string) string {
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

func generateBase58(length int) (string, error) {
	alphabet := []rune(base58Alphabet)
	result := make([]rune, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[n.Int64()]
	}
	return string(result), nil
}

func generateHex(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
