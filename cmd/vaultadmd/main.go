package main

import (
	"context"
	"errors"
	"fmt"
	"net"

	vaultapi "github.com/hashicorp/vault/api"
	"google.golang.org/grpc"
	hproto "google.golang.org/grpc/health/grpc_health_v1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	vproto "github.com/ZentriaMC/vaultadm/pkg/proto"
)

func main() {
	if err := entrypoint(); err != nil {
		panic(err)
	}
}

func entrypoint() (err error) {
	var lis net.Listener
	if lis, err = net.Listen("tcp", fmt.Sprintf(":%d", 50059)); err != nil {
		return
	}

	svc := &svc{}
	if svc.client, err = vaultapi.NewClient(vaultapi.DefaultConfig()); err != nil {
		return
	}

	srv := grpc.NewServer()
	vproto.RegisterManagerServer(srv, svc)
	hproto.RegisterHealthServer(srv, svc)

	if err = srv.Serve(lis); err != nil {
		return
	}

	return
}

type svc struct {
	vproto.UnimplementedManagerServer
	hproto.UnimplementedHealthServer

	client *vaultapi.Client
}

func (s *svc) UnsealPortion(ctx context.Context, req *vproto.UnsealRequest) (resp *vproto.UnsealResponse, err error) {
	var res *vaultapi.SealStatusResponse
	res, err = s.client.Sys().UnsealWithOptionsWithContext(ctx, &vaultapi.UnsealOpts{
		Key:     req.GetPortion(),
		Reset:   req.GetReset_(),
		Migrate: req.GetMigrate(),
	})
	if err != nil {
		return
	}

	resp = &vproto.UnsealResponse{
		Sealed:    res.Sealed,
		Threshold: int32(res.T),
		Shares:    int32(res.N),
		Progress:  int32(res.Progress),
	}
	return
}

func (s *svc) Seal(ctx context.Context, req *vproto.SealRequest) (resp *emptypb.Empty, err error) {
	err = s.client.Sys().SealWithContext(ctx)
	return &emptypb.Empty{}, err
}

func (s *svc) ObtainRootToken(ctx context.Context, req *vproto.RootTokenRequest) (resp *vproto.RootTokenResponse, err error) {
	var secret *vaultapi.Secret
	secret, err = s.client.Auth().Token().CreateWithContext(ctx, &vaultapi.TokenCreateRequest{
		TTL:      fmt.Sprintf("%ds", req.GetTtl()),
		NoParent: req.GetOrphan(),
	})
	if err != nil {
		return
	}

	resp = &vproto.RootTokenResponse{
		Token:         secret.Auth.ClientToken,
		Accessor:      secret.Auth.Accessor,
		Policies:      secret.Auth.Policies,
		TokenPolicies: secret.Auth.TokenPolicies,
		Metadata:      secret.Auth.Metadata,
		LeaseDuration: uint64(secret.Auth.LeaseDuration),
		Renewable:     secret.Auth.Renewable,
		Orphan:        secret.Auth.Orphan,
	}
	return resp, nil
}

func (s *svc) Check(ctx context.Context, req *hproto.HealthCheckRequest) (*hproto.HealthCheckResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *svc) Watch(req *hproto.HealthCheckRequest, srv hproto.Health_WatchServer) error {
	return errors.New("not implemented")
}
