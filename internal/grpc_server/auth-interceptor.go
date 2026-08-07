package grpc_server

import (
	"context"
	"errors"

	"github.com/SergeyRG/shortener/internal/auth"
	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewAuthInterceptor(cfg config.Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		publicMethods := map[string]bool{
			"/shortener.v1.ShortenerService/ExpandURL": true,
		}

		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		tokenString := ""
		md, present := metadata.FromIncomingContext(ctx)
		if !present {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		jwt, exist := md["auth_token"]
		if exist && len(jwt) > 0 {
			tokenString = jwt[0]
		}

		userID, err := auth.ProccessToken(tokenString, []byte(cfg.SecretKey))

		var e auth.ErrNewTokenRequerd
		if errors.As(err, &e) {
			authMD := metadata.Pairs("auth_token", e.NewToken)
			grpc.SetHeader(ctx, authMD)
			err = nil
		}

		if err != nil {
			logging.Logger.Error("auth error", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		if userID == "" {
			return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
		}

		outgoingCtx := auth.ContextWithUserID(ctx, userID)

		return handler(outgoingCtx, req)
	}
}
