package grpc_server

import (
	authv1 "auth/gen/auth/v1"
	"auth/internal/cache"
	"auth/internal/errs"
	"auth/internal/repositories"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	repo *repositories.Repository
	c    *cache.Cache
}

func BuildAuthGrpcServer(repo *repositories.Repository, c *cache.Cache) *AuthServer {
	return &AuthServer{
		repo: repo,
		c:    c,
	}
}

func (a *AuthServer) VerifyUserExists(ctx context.Context, req *authv1.VerifyUserExistsRequest) (*authv1.VerifyUserExistsResponse, error) {
	if req.Email == "" || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and Id must be provided")
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "The id provided should be parseable to uuid")
	}
	err = a.repo.CheckIfUserExists(ctx, id, req.Email)
	if err != nil {
		if errors.Is(err, errs.ERR_EMAIL_NO_EXISTS) {
			// return nil, status.Error(codes.NotFound, fmt.Sprintf("No user exists with id of %s and email of %s", req.Id, req.Email))

			return &authv1.VerifyUserExistsResponse{Exists: false}, nil
		}
		return nil, status.Error(codes.Unknown, "Something went wrong")
	}

	return &authv1.VerifyUserExistsResponse{
		Exists: true,
	}, nil
}

func (a *AuthServer) AutheticateUser(ctx context.Context, req *authv1.AuthenticateUserRequest) (*authv1.AuthenticateUserResponse, error) {
	sessionId := req.SessionID
	if sessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "SessionId can not be empty")
	}
	user, allowed := a.c.GetUser(ctx, sessionId)

	return &authv1.AuthenticateUserResponse{
		Allow: allowed,
		User: &authv1.MinimalUserStruct{
			UserId: user.UserId.String(),
			Email:  user.Email,
		},
	}, nil
}
