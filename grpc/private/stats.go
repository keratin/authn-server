package private

import (
	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	"golang.org/x/net/context"
)

type statsServer struct {
	app *app.App

	// SECURITY: ensure that both ConstantTimeCompare operations are run, so that a
	// timing attack may not verify a correct username without a correct password.
	matcher func(username, password string) bool
}

func (ss statsServer) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return basicAuthCheck(ctx, ss.matcher)
}

func (ss statsServer) ServiceStats(ctx context.Context, req *authnpb.ServiceStatsRequest) (*authnpb.ServiceStatsResponse, error) {
	daily, err := ss.app.Actives.ActivesByDay()
	if err != nil {
		panic(err)
	}

	weekly, err := ss.app.Actives.ActivesByWeek()
	if err != nil {
		panic(err)
	}

	monthly, err := ss.app.Actives.ActivesByMonth()
	if err != nil {
		panic(err)
	}

	return &authnpb.ServiceStatsResponse{
		Actives: &authnpb.ServiceStatsResponseActiveStats{
			Daily:   toMapStringInt64(daily),
			Weekly:  toMapStringInt64(weekly),
			Monthly: toMapStringInt64(monthly),
		},
	}, nil
}

func toMapStringInt64(m map[string]int) map[string]int64 {
	out := map[string]int64{}
	for k, v := range m {
		out[k] = int64(v)
	}
	return out
}
