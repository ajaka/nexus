package grpc_client

import authv1 "gateway/gen/auth/v1"

type GrpcClient struct {
	Client authv1.AuthServiceClient
}
