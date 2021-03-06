package main

import (
	"context"
	"os"

	"github.com/gidyon/micro"
	"github.com/gidyon/micro/utils/healthcheck"

	pusher_app "github.com/gidyon/services/internal/messaging/pusher"

	"github.com/gidyon/services/pkg/api/messaging/pusher"
	"github.com/gidyon/services/pkg/utils/errs"

	"github.com/gidyon/micro/pkg/config"
	app_grpc_middleware "github.com/gidyon/micro/pkg/grpc/middleware"
)

func main() {
	ctx := context.Background()

	// Read config
	cfg, err := config.New(config.FromFile)
	errs.Panic(err)

	// Create service instance
	app, err := micro.NewService(ctx, cfg, nil)
	errs.Panic(err)

	// Recovery middleware
	recoveryUIs, recoverySIs := app_grpc_middleware.AddRecovery()
	app.AddGRPCUnaryServerInterceptors(recoveryUIs...)
	app.AddGRPCStreamServerInterceptors(recoverySIs...)

	// Readiness health check
	app.AddEndpoint("/api/pusher/health/ready", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeReadiness,
		AutoMigrator: func() error { return nil },
	}))

	// Liveness health check
	app.AddEndpoint("/api/pusher/health/live", healthcheck.RegisterProbe(&healthcheck.ProbeOptions{
		Service:      app,
		Type:         healthcheck.ProbeLiveNess,
		AutoMigrator: func() error { return nil },
	}))

	app.Start(ctx, func() error {
		// Create pusher API
		pusherAPI, err := pusher_app.NewPushMessagingServer(ctx, &pusher_app.Options{
			Logger:        app.Logger(),
			JWTSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
			FCMServerKey:  os.Getenv("FCM_SERVER_KEY"),
		})
		errs.Panic(err)

		pusher.RegisterPushMessagingServer(app.GRPCServer(), pusherAPI)
		pusher.RegisterPushMessagingHandler(ctx, app.RuntimeMux(), app.ClientConn())

		return nil
	})
}
