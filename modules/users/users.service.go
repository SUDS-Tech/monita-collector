package users

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/bastion-framework/bast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/SUDS-Tech/monita-collector/shared/guards"
	apperr "github.com/SUDS-Tech/monita-collector/shared/errors"
)

type Service struct {
	repo      *repo
	jwtSecret []byte
}

func newService(r *repo, jwtSecret []byte) *Service {
	return &Service{repo: r, jwtSecret: jwtSecret}
}

// OrgDetails is returned by FindOrgByStripeCustomer.
// It lives here so billing.OrgProvider can reference it without a circular import.
type OrgDetails struct {
	ID   string
	Name string
	Slug string
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.createUser(ctx, req.Email, req.Name, string(hash))
	if err != nil {
		if isUniqueViolation(err) {
			return nil, bast.ErrConflict(apperr.CodeEmailTaken, "email already registered")
		}
		return nil, err
	}

	org, err := s.repo.createOrganization(ctx, req.Name, toSlug(req.Name, user.ID), user.ID)
	if err != nil {
		return nil, err
	}

	token, err := guards.SignToken(s.jwtSecret, user.ID.String(), org.ID.String(), user.Email)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	row, err := s.repo.getUserForLogin(ctx, req.Email)
	if err != nil {
		return nil, bast.ErrUnauthorized(apperr.CodeInvalidCredentials, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(req.Password)); err != nil {
		return nil, bast.ErrUnauthorized(apperr.CodeInvalidCredentials, "invalid credentials")
	}

	org, err := s.repo.getOrganizationByOwnerID(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	token, err := guards.SignToken(s.jwtSecret, row.ID.String(), org.ID.String(), row.Email)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token}, nil
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*MeResponse, error) {
	user, err := s.repo.getUserByID(ctx, userID)
	if err != nil {
		return nil, bast.ErrNotFound("USER_NOT_FOUND", "user not found")
	}

	org, err := s.repo.getOrganizationByOwnerID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &MeResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
		OrgID: org.ID.String(),
	}, nil
}

// FindOrgByStripeCustomer and SetSubscription satisfy billing.OrgProvider (duck typing).
// Implemented when the billing module is built.

func (s *Service) FindOrgByStripeCustomer(_ context.Context, _ string) (*OrgDetails, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) SetSubscription(_ context.Context, _, _, _, _ string) error {
	return errors.New("not implemented")
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toSlug(name string, userID uuid.UUID) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
			b.WriteRune('-')
		}
	}
	slug := strings.TrimRight(b.String(), "-")
	if slug == "" {
		slug = "org"
	}
	uid := strings.ReplaceAll(userID.String(), "-", "")
	return slug + "-" + uid[:8]
}
