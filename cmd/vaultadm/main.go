package main

import (
	"context"
	"os"
	"os/signal"

	sshgrpc "github.com/ZentriaMC/grpc-ssh/pkg/client"
	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	vproto "github.com/ZentriaMC/vaultadm/pkg/proto"
)

func main() {
	if err := configureLogging(true); err != nil {
		panic(err)
	}
	defer func() { _ = zap.L().Sync() }()

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	sshDialer, err := sshgrpc.NewDialer(sshgrpc.SSHConnectionDetails{
		Hostname:    "127.0.0.1",
		EnableAgent: true,
	})
	if err != nil {
		panic(err)
	}

	conn, err := grpc.DialContext(ctx, "vaultadm", []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(sshDialer.Dialer()),
	}...)
	if err != nil {
		panic(err)
	}

	client := vproto.NewManagerClient(conn)
	//res, err := client.Seal(ctx, &vproto.SealRequest{})
	res, err := client.ObtainRootToken(ctx, &vproto.RootTokenRequest{
		Ttl: 60,
	})
	if err != nil {
		panic(err)
	}

	spew.Dump(res)
}

func configureLogging(debugMode bool) (err error) {
	var cfg zap.Config

	if debugMode {
		cfg = zap.NewDevelopmentConfig()
		cfg.Level.SetLevel(zapcore.DebugLevel)
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.Development = false
	} else {
		cfg = zap.NewProductionConfig()
		cfg.Level.SetLevel(zapcore.InfoLevel)
	}

	cfg.OutputPaths = []string{
		"stderr",
	}

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	_ = zap.ReplaceGlobals(logger)
	return
}
