package guards

import (
	"strings"
	"time"

	"github.com/bastion-framework/bast"
	"github.com/golang-jwt/jwt/v5"

	apperr "github.com/SUDS-Tech/monita-collector/shared/errors"
)

// UserClaims is set in request context on successful auth.
// Key: "user". Handlers read it via ctx.MustGet("user").(*guards.UserClaims).
type UserClaims struct {
	UserID string
	OrgID  string
	Email  string
}


type SessionClaims struct {
	jwt.RegisteredClaims
	OrgID string `json:"org_id"`
	Email string `json:"email"`
}

type SessionAuthGuard struct {
	secret []byte
}

func NewSessionAuth(secret string) *SessionAuthGuard {
	return &SessionAuthGuard{secret: []byte(secret)}
}

func (g *SessionAuthGuard) SecurityScheme() bast.SecurityScheme {
	return bast.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT issued at login. Pass as: Authorization: Bearer <token>",
	}
}

func (g *SessionAuthGuard) Check(ctx *bast.Ctx) error {
	header := ctx.Header("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return bast.ErrUnauthorized(apperr.CodeInvalidCredentials, "invalid or missing token")
	}
	raw := strings.TrimPrefix(header, "Bearer ")

	claims := &SessionClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, bast.ErrUnauthorized(apperr.CodeInvalidCredentials, "invalid signing method")
		}
		return g.secret, nil
	})
	if err != nil || !token.Valid {
		return bast.ErrUnauthorized(apperr.CodeInvalidCredentials, "invalid or missing token")
	}

	ctx.Set("user", &UserClaims{
		UserID: claims.Subject,
		OrgID:  claims.OrgID,
		Email:  claims.Email,
	})
	return nil
}


func SignToken(secret []byte, userID, orgID, email string) (string, error) {
	claims := SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		OrgID: orgID,
		Email: email,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}