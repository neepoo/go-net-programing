package main

import (
	"context"
	"fmt"
	"sync"

	pb "go-network/chapter12/housework/v1"
)

type Rosie struct {
	mu     sync.Mutex
	chores []*pb.Chore
	pb.UnimplementedRobotMaidServer
}

func NewRosie() *Rosie {
	return &Rosie{chores: make([]*pb.Chore, 0)}
}

func (r *Rosie) Add(_ context.Context, chores *pb.Chores) (*pb.Response, error) {
	r.mu.Lock()
	r.chores = append(r.chores, chores.Chores...)
	r.mu.Unlock()

	return &pb.Response{
		Message: "ok",
	}, nil
}

func (r *Rosie) Complete(_ context.Context, req *pb.CompleteRequest) (*pb.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.chores == nil || req.ChoreNumber < 1 ||
		int(req.ChoreNumber) > len(r.chores) {
		return nil, fmt.Errorf("chore %d not found", req.ChoreNumber)
	}
	r.chores[req.ChoreNumber-1].Complete = true

	return &pb.Response{Message: "ok"}, nil
}

func (r *Rosie) List(_ context.Context, _ *pb.Empty) (*pb.Chores, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.chores == nil {
		r.chores = make([]*pb.Chore, 0)
	}
	return &pb.Chores{Chores: r.chores}, nil
}
