package private

import (
	"github.com/keratin/authn-server/grpc/internal/errors"
	"google.golang.org/grpc/codes"

	"github.com/keratin/authn-server/api"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/services"
	"golang.org/x/net/context"
)

type securedServer struct {
	app *api.App

	// SECURITY: ensure that both ConstantTimeCompare operations are run, so that a
	// timing attack may not verify a correct username without a correct password.
	matcher func(username, password string) bool
}

func (ss securedServer) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return basicAuthCheck(ctx, ss.matcher)
}

func (ss securedServer) GetAccount(ctx context.Context, req *authnpb.GetAccountRequest) (*authnpb.GetAccountResponseEnvelope, error) {

	account, err := services.AccountGetter(ss.app.AccountStore, int(req.GetId()))
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.NotFound).Err()
		}
		panic(err)
	}

	return &authnpb.GetAccountResponseEnvelope{
		Result: &authnpb.GetAccountResponse{
			Id:       int64(account.ID),
			Username: account.Username,
			Locked:   account.Locked,
			Deleted:  account.DeletedAt != nil,
		},
	}, nil
}

func (ss securedServer) UpdateAccount(ctx context.Context, req *authnpb.UpdateAccountRequest) (*authnpb.UpdateAccountResponse, error) {
	err := services.AccountUpdater(ss.app.AccountStore, ss.app.Config, int(req.GetId()), req.GetUsername())
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			if fe[0].Message == services.ErrNotFound {
				return nil, errors.ToStatusErrorWithDetails(fe, codes.NotFound).Err()
			}

			return nil, errors.ToStatusErrorWithDetails(fe, codes.FailedPrecondition).Err()
		}
		panic(err)
	}

	return &authnpb.UpdateAccountResponse{}, nil
}

func (ss securedServer) LockAccount(ctx context.Context, req *authnpb.LockAccountRequest) (*authnpb.LockAccountResponse, error) {
	err := services.AccountLocker(ss.app.AccountStore, ss.app.RefreshTokenStore, int(req.GetId()))
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.NotFound).Err()
		}

		panic(err)
	}

	return &authnpb.LockAccountResponse{}, nil
}

func (ss securedServer) UnlockAcount(ctx context.Context, req *authnpb.UnlockAccountRequest) (*authnpb.UnlockAccountResponse, error) {
	err := services.AccountUnlocker(ss.app.AccountStore, int(req.GetId()))
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.NotFound).Err()
		}

		panic(err)
	}

	return &authnpb.UnlockAccountResponse{}, nil
}

func (ss securedServer) ArchiveAccount(ctx context.Context, req *authnpb.ArchiveAccountRequest) (*authnpb.ArchiveAccountResponse, error) {
	err := services.AccountArchiver(ss.app.AccountStore, ss.app.RefreshTokenStore, int(req.GetId()))
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.FailedPrecondition).Err()
		}
		panic(err)
	}

	return &authnpb.ArchiveAccountResponse{}, nil
}

func (ss securedServer) ImportAccount(ctx context.Context, req *authnpb.ImportAccountRequst) (*authnpb.ImportAccountResponseEnvelope, error) {
	account, err := services.AccountImporter(
		ss.app.AccountStore,
		ss.app.Config,
		req.GetUsername(),
		req.GetPassword(),
		req.GetLocked(),
	)
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.FailedPrecondition).Err()
		}
		panic(err)
	}

	return &authnpb.ImportAccountResponseEnvelope{
		Result: &authnpb.ImportAccountResponse{
			Id: int64(account.ID),
		},
	}, nil
}

func (ss securedServer) ExpirePassword(ctx context.Context, req *authnpb.ExpirePasswordRequest) (*authnpb.ExpirePasswordResponse, error) {
	err := services.PasswordExpirer(ss.app.AccountStore, ss.app.RefreshTokenStore, int(req.GetId()))
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.NotFound).Err()
		}
		panic(err)
	}

	return &authnpb.ExpirePasswordResponse{}, nil
}
