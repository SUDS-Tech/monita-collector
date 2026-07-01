package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	dbsqlc "github.com/SUDS-Tech/monita-collector/internal/db/sqlc"
)

type repo struct {
	q    *dbsqlc.Queries
	pool *pgxpool.Pool
}

func newRepo(pool *pgxpool.Pool) *repo {
	return &repo{q: dbsqlc.New(pool), pool: pool}
}

func (r *repo) createUser(ctx context.Context, email, name, passwordHash string) (dbsqlc.CreateUserRow, error) {
	return r.q.CreateUser(ctx, dbsqlc.CreateUserParams{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
	})
}

func (r *repo) getUserForLogin(ctx context.Context, email string) (dbsqlc.GetUserForLoginRow, error) {
	return r.q.GetUserForLogin(ctx, email)
}

func (r *repo) getUserByID(ctx context.Context, id uuid.UUID) (dbsqlc.GetUserByIDRow, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *repo) createOrganization(ctx context.Context, name, slug string, ownerID uuid.UUID) (dbsqlc.CreateOrganizationRow, error) {
	return r.q.CreateOrganization(ctx, dbsqlc.CreateOrganizationParams{
		Name:    name,
		Slug:    slug,
		OwnerID: ownerID,
	})
}

func (r *repo) getOrganizationByOwnerID(ctx context.Context, ownerID uuid.UUID) (dbsqlc.GetOrganizationByOwnerIDRow, error) {
	return r.q.GetOrganizationByOwnerID(ctx, ownerID)
}
