// This file contains the interfaces for the repository layer.
// The repository layer is responsible for interacting with the database.
// For testing purpose we will generate mock implementations of these
// interfaces using mockgen. See the Makefile for more information.
package repository

import (
	"context"
)

type RepositoryInterface interface {
	GetTestById(ctx context.Context, input GetTestByIdInput) (output GetTestByIdOutput, err error)
	InsertEstate(ctx context.Context, input EstateRequest) (output EstateResponse, err error)
	InsertTree(ctx context.Context, input TreeRequest) (output TreeResponse, err error)
	ValidateEstateRequest(ctx context.Context, input EstateRequest) (err error)
	ValidateTreeRequest(ctx context.Context, estateId string, input TreeRequest) (err error)
	GetEstateStats(ctx context.Context, estateId string) (EstateStats, error)
	GetEstateById(ctx context.Context, id string) (EstateData, error)
	GetTreesByEstateId(ctx context.Context, estateId string) ([]Tree, error)
}
