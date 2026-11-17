package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID  int64  `json:"uid"`
	Role    string `json:"role"`
	StoreID *int64 `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	expire time.Duration
}

func NewManager(secret string, expireHours int) *Manager {
	if expireHours <= 0 {
		expireHours = 72
	}
	return &Manager{secret: []byte(secret), expire: time.Duration(expireHours) * time.Hour}
}

func (m *Manager) Issue(userID int64, role string, storeID *int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		Role:    role,
		StoreID: storeID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expire)),
			Issuer:    "lanping",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

func (m *Manager) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
