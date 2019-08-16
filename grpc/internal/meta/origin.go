package meta

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type matchedDomainKey int

// audienceMatcher figures the expected audience of JWT from Origin or Referer headers if the request
// comes from REST, otherwise uses AuthN isser URL as the audience.
func audienceMatcher(app *app.App) grpc.UnaryServerInterceptor {
	var validDomains []string
	for _, d := range app.Config.ApplicationDomains {
		validDomains = append(validDomains, d.String())
	}
	logger := log.WithFields(log.Fields{"validDomains": validDomains})
	gRPCIssuer := route.ParseDomain(fmt.Sprintf("%s:%s", app.Config.AuthNURL.Host, app.Config.AuthNURL.Port()))

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		origin := inferOrigin(ctx)

		// If we don't receive an origin from inferOrigin, then we can
		// assume the request is coming from gRPC-only client, and not
		// through the gRPC-Gateway, because it couldn't have passed through the
		// Origin-validation on the first router.
		if origin == "" {
			ctx = setAudience(ctx, gRPCIssuer)
			return handler(ctx, req)
		}

		domain := route.FindDomain(origin, app.Config.ApplicationDomains)
		if domain == nil {
			// This shouldn't be reachable? Log then return
			logger.Debug("Could not infer request origin, and request not treated as gRPC request")
			err := status.New(codes.PermissionDenied, `No valid Origin domain found`).Err()
			app.Reporter.ReportGRPCError(err, info, req)
			return nil, err
		}

		ctx = setAudience(ctx, *domain)
		return handler(ctx, req)
	}
}

// MatchedDomain will retrieve from the http.Request's Context the domain that satisfied
// OriginSecurity.
func MatchedDomain(ctx context.Context) *route.Domain {
	d, ok := ctx.Value(matchedDomainKey(0)).(route.Domain)
	if ok {
		return &d
	}
	return nil
}

func setAudience(ctx context.Context, domain route.Domain) context.Context {
	return context.WithValue(ctx, matchedDomainKey(0), domain)
}

func inferOrigin(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	origins := md.Get(getGatewayHeaderKey("Origin"))
	if len(origins) != 0 {
		return origins[0]
	}

	// If and only if the origin header is unset we can infer that this is a same-origin request
	// (i.e we trust browsers to behave this way), then we use the Referer header to discover the domain
	origins = md.Get(getGatewayHeaderKey("Referer"))
	if len(origins) != 0 {
		return origins[0]
	}
	return ""
}

func getGatewayHeaderKey(key string) (gkey string) {
	gkey, _ = runtime.DefaultHeaderMatcher(key)
	return
}
