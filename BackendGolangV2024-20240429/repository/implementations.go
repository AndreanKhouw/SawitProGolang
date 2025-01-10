package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

func (r *Repository) GetTestById(ctx context.Context, input GetTestByIdInput) (output GetTestByIdOutput, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT name FROM test WHERE id = $1", input.Id).Scan(&output.Name)
	if err != nil {
		return
	}
	return
}

func (r *Repository) InsertEstate(ctx context.Context, input EstateRequest) (EstateResponse, error) {
	var id uuid.UUID
	query := "INSERT INTO estate (length, width) VALUES ($1, $2) RETURNING id"
	err := r.Db.QueryRowContext(ctx, query, input.Length, input.Width).Scan(&id)
	if err != nil {
		log.Printf("Error inserting estate: %v\n", err)
		return EstateResponse{}, err // Return an empty EstateResponse and the error
	}

	response := EstateResponse{
		Id: id,
	}

	return response, nil
}

func (r *Repository) InsertTree(ctx context.Context, input TreeRequest) (TreeResponse, error) {
	var id uuid.UUID
	query := "INSERT INTO tree (estateid,x,y,height) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.Db.QueryRowContext(ctx, query, input.EstateId, input.X, input.Y, input.Height).Scan(&id)
	if err != nil {
		log.Printf("Error inserting Tree: %v\n", err)
		return TreeResponse{}, err // Return an empty EstateResponse and the error
	}

	response := TreeResponse{
		Id: id,
	}

	return response, nil
}

func (r *Repository) ValidateTreeRequest(ctx context.Context, estateId string, input TreeRequest) error {
	var length, width int
	query := "SELECT length, width FROM estate WHERE id = $1"
	err := r.Db.QueryRowContext(ctx, query, estateId).Scan(&length, &width)
	if err != nil {
		return fmt.Errorf("estate not found or database error: %v", err)
	}
	if input.X > length || input.X <= 0 {
		return fmt.Errorf("x (%d) exceeds estate length (%d)", input.X, length)
	}
	if input.Y > width || input.Y <= 0 {
		return fmt.Errorf("y (%d) exceeds estate width (%d)", input.Y, width)
	}

	if input.Height > 30 {
		return fmt.Errorf("height (%d) exceeds the maximum allowed value (30)", input.Height)
	}

	return nil
}

func (r *Repository) ValidateEstateRequest(ctx context.Context, input EstateRequest) error {

	if input.Length <= 0 {
		return fmt.Errorf("length (%d) can not less than 0 ", input.Length)
	}

	if input.Width <= 0 {
		return fmt.Errorf("width (%d) can not less than 0", input.Width)
	}
	return nil
}

func (r *Repository) GetEstateStats(ctx context.Context, estateId string) (EstateStats, error) {
	var count, max, min int
	var median float64

	// Check if estate exists
	var estateExists bool
	err := r.Db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM estate WHERE id = $1)", estateId).Scan(&estateExists)
	if err != nil || !estateExists {
		return EstateStats{}, fmt.Errorf("estate not found")
	}

	// Query for count, max, and min
	err = r.Db.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(height), MIN(height)
		FROM tree
		WHERE estateId = $1
	`, estateId).Scan(&count, &max, &min)
	if err != nil {
		return EstateStats{}, err
	}
	if count == 0 {
		return EstateStats{
			Count:     0,
			MaxHeight: 0,
			MinHeight: 0,
			Median:    0,
		}, nil
	}
	// Query for median
	err = r.Db.QueryRowContext(ctx, `
		SELECT PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY height)
		FROM tree
		WHERE estateId = $1
	`, estateId).Scan(&median)
	if err != nil {
		median = 0.0
	}

	return EstateStats{
		Count:     count,
		MaxHeight: max,
		MinHeight: min,
		Median:    median,
	}, nil
}

func (r *Repository) GetEstateById(ctx context.Context, id string) (EstateData, error) {
	var estate EstateData
	err := r.Db.QueryRowContext(ctx, "SELECT id, length, width FROM estate WHERE id = $1", id).
		Scan(&estate.Id, &estate.Length, &estate.Width)
	if err != nil {
		return EstateData{}, fmt.Errorf("estate not found")
	}
	return estate, nil
}

func (r *Repository) GetTreesByEstateId(ctx context.Context, estateId string) ([]Tree, error) {
	rows, err := r.Db.QueryContext(ctx, `
		SELECT x, y, height
		FROM tree
		WHERE estateId = $1
	`, estateId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trees []Tree
	for rows.Next() {
		var tree Tree
		if err := rows.Scan(&tree.X, &tree.Y, &tree.Height); err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}
	return trees, nil
}
