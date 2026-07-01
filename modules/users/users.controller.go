package users

import (
	"github.com/bastion-framework/bast"
	"github.com/google/uuid"

	"github.com/SUDS-Tech/monita-collector/shared/guards"
	apperr "github.com/SUDS-Tech/monita-collector/shared/errors"
)

type controller struct {
	svc          *Service
	sessionGuard bast.Guard
}

func newController(svc *Service, sessionGuard bast.Guard) *controller {
	return &controller{svc: svc, sessionGuard: sessionGuard}
}

func (c *controller) Routes() []bast.Route {
	return []bast.Route{
		bast.POST("/register", c.Register, bast.WithDoc(bast.Doc{
			Summary: "Register a new user",
			Tags:    []string{"Users"},
			Body:    bast.Body[RegisterRequest](),
			Returns: bast.Returns{201: bast.Body[AuthResponse]()},
		})),
		bast.POST("/login", c.Login, bast.WithDoc(bast.Doc{
			Summary: "Login and receive a JWT",
			Tags:    []string{"Users"},
			Body:    bast.Body[LoginRequest](),
			Returns: bast.Returns{200: bast.Body[AuthResponse]()},
		})),
		bast.GET("/me", c.Me, bast.WithGuards(c.sessionGuard), bast.WithDoc(bast.Doc{
			Summary: "Get current user profile",
			Tags:    []string{"Users"},
			Returns: bast.Returns{200: bast.Body[MeResponse]()},
		})),
	}
}

func (c *controller) Register(ctx *bast.Ctx) bast.Response {
	var req RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Error(err)
	}
	resp, err := c.svc.Register(ctx.Context(), req)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.Created(resp)
}

func (c *controller) Login(ctx *bast.Ctx) bast.Response {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Error(err)
	}
	resp, err := c.svc.Login(ctx.Context(), req)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(resp)
}

func (c *controller) Me(ctx *bast.Ctx) bast.Response {
	claims := ctx.MustGet("user").(*guards.UserClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return ctx.Error(bast.ErrUnauthorized(apperr.CodeInvalidCredentials, "invalid user ID"))
	}
	me, err := c.svc.Me(ctx.Context(), userID)
	if err != nil {
		return ctx.Error(err)
	}
	return ctx.OK(me)
}