package mock

import (
	"context"
	bosonFunc "github.com/boson-project/func"
)

type Deployer struct {
	DeployInvoked bool
	DeployFn      func(bosonFunc.Function) error
}

func NewDeployer() *Deployer {
	return &Deployer{
		DeployFn: func(bosonFunc.Function) error { return nil },
	}
}

func (i *Deployer) Deploy(ctx context.Context, f bosonFunc.Function) error {
	i.DeployInvoked = true
	return i.DeployFn(f)
}
