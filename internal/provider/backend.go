package provider

import (
	"context"

	"github.com/ImTheCurse/ConflowCI/internal/provider/github/pb"
)

type RepositoryReader interface {
	Clone(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error)
	Fetch(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error)
	CreateWorkTree(context.Context, *pb.WorkTreeRequest) (*pb.SyncResponse, error)
	RemoveWorkTree(context.Context, *pb.WorkTreeRequest) (*pb.SyncResponse, error)
}
