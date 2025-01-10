// This file contains types that are used in the repository layer.
package repository

import "github.com/google/uuid"

type GetTestByIdInput struct {
	Id string
}

type GetTestByIdOutput struct {
	Name string
}

type EstateRequest struct {
	Length int
	Width  int
}

type EstateResponse struct {
	Id uuid.UUID
}

type TreeRequest struct {
	EstateId string
	Height   int
	X        int
	Y        int
}

type TreeResponse struct {
	Id uuid.UUID
}
type EstateStats struct {
	Count     int     `json:"count"`
	MaxHeight int     `json:"max"`
	MinHeight int     `json:"min"`
	Median    float64 `json:"median"`
}

type Tree struct {
	X      int
	Y      int
	Height int
}

type EstateData struct {
	Id     uuid.UUID
	Length int
	Width  int
}
