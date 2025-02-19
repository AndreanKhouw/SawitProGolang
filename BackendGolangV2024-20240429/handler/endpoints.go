package handler

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// This is just a test endpoint to get you started. Please delete this endpoint.
// (GET /hello)
func (s *Server) GetHello(ctx echo.Context, params generated.GetHelloParams) error {
	var resp generated.HelloResponse
	resp.Message = fmt.Sprintf("Hello User %d", params.Id)
	return ctx.JSON(http.StatusOK, resp)
}

type EstateInput struct {
	Length int32 `json:"length"`
	Width  int32 `json:"width"`
}

func (s *Server) PostEstate(ctx echo.Context) error {
	var input repository.EstateRequest
	if ctx.Request().ContentLength == 0 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Request body is missing"})
	}
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if err := s.Repository.ValidateEstateRequest(context.Background(), input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	id, err := s.Repository.InsertEstate(ctx.Request().Context(), input)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert estate"})
	}
	uuid := id.Id.String()
	if uuid != "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"id": uuid,
		})
	} else {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert estate"})
	}
}

func (s *Server) PostTree(ctx echo.Context, estateId string) error {
	if _, err := uuid.Parse(estateId); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid estate ID")
	}

	var req repository.TreeRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	// Debugging: Print incoming data
	fmt.Println("Tree coordinates:", req.X, req.Y)
	if err := s.Repository.ValidateTreeRequest(context.Background(), estateId, req); err != nil {

		// Debugging: Print incoming data
		fmt.Println("Gotchhaaaaaa shoulbe right")
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// Interact with the repository to insert the tree
	response, err := s.Repository.InsertTree(context.Background(), repository.TreeRequest{
		EstateId: estateId,
		X:        req.X,
		Y:        req.Y,
		Height:   req.Height,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to add tree"})
	}

	uuid := response.Id.String()
	if uuid != "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"id": uuid,
		})
	} else {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert Tree"})
	}
}

func (s *Server) GetStats(ctx echo.Context, id string) error {
	stats, err := s.Repository.GetEstateStats(context.Background(), id)
	if err != nil {
		if err.Error() == "estate not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "estate not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	if math.IsNaN(stats.Median) {
		stats.Median = 0.0 // Or set to NaN, depending on your preference
	}

	response := map[string]interface{}{
		"count":  stats.Count,
		"max":    stats.MaxHeight,
		"min":    stats.MinHeight,
		"median": stats.Median,
	}

	return ctx.JSON(http.StatusOK, response)
}

func (s *Server) GetEstateIdDronePlan(ctx echo.Context, id string) error {
	estate, err := s.Repository.GetEstateById(context.Background(), id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "estate not found"})
	}

	trees, err := s.Repository.GetTreesByEstateId(context.Background(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	totalElevation := calculateTotalElevation(trees)

	totalHorizontal := ((estate.Length * estate.Width) - 1) * 10
	totalDistance := totalHorizontal + totalElevation + 2

	response := map[string]interface{}{
		"distance": totalDistance,
	}

	return ctx.JSON(http.StatusOK, response)
}

// Sort trees in a zigzag pattern
func sortTrees(trees []repository.Tree) {
	sort.SliceStable(trees, func(i, j int) bool {
		if trees[i].Y == trees[j].Y {
			if trees[i].Y%2 == 0 {
				return trees[i].X < trees[j].X // Even rows: left to right
			}
			return trees[i].X > trees[j].X // Odd rows: right to left
		}
		return trees[i].Y < trees[j].Y // Top to bottom
	})
}

func calculateTotalElevation(trees []repository.Tree) int {
	totalElevation := 0

	// Sort trees in zigzag order
	sortTrees(trees)

	for i, tree := range trees {
		if i == 0 {
			// First tree: add its height
			totalElevation += tree.Height
		} else {
			// Calculate the difference in X and Y
			prevTree := trees[i-1]
			diffX := math.Abs(float64(tree.X - prevTree.X))
			diffY := math.Abs(float64(tree.Y - prevTree.Y))

			// Check if trees are adjacent
			if diffX <= 1 && diffY <= 1 {
				// Adjacent: Add the height difference
				totalElevation += int(math.Abs(float64(tree.Height - prevTree.Height)))
			} else {
				// Not adjacent: Add both heights (current and next tree)
				totalElevation += prevTree.Height + tree.Height
			}
		}
	}
	totalElevation += trees[len(trees)-1].Height

	return totalElevation
}

func (s *Server) GetEstateIdDronePlanWithMaxDistance(ctx echo.Context, id string, params generated.GetEstateIdDronePlanWithMaxDistanceParams) error {
	estate, err := s.Repository.GetEstateById(context.Background(), id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "estate not found"})
	}

	trees, err := s.Repository.GetTreesByEstateId(context.Background(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	// Calculate total elevation and horizontal distance
	totalElevation := calculateTotalElevation(trees)
	// Ensure each horizontal movement is multiplied by 10 meters
	totalHorizontal := ((estate.Length * estate.Width) - 1) * 10
	totalDistance := totalHorizontal + totalElevation + 2

	maxDistance := params.MaxDistance
	if maxDistance > totalDistance {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid max_distance"})
	}

	// If max_distance is provided and is less than totalDistance
	if maxDistance > 0 && totalDistance > maxDistance {
		landingPoint := calculateLandingPlot(trees, maxDistance, totalHorizontal, estate.Width)
		response := map[string]interface{}{
			"distance": maxDistance,
			"rest":     landingPoint,
		}
		return ctx.JSON(http.StatusOK, response)
	} else {
		// Otherwise, return the last plot coordinates
		landingPoint := map[string]int{
			"x": estate.Length,
			"y": estate.Width,
		}

		response := map[string]interface{}{
			"distance":      maxDistance,
			"landing_point": landingPoint,
		}
		return ctx.JSON(http.StatusOK, response)
	}
}
func calculateLandingPlot(trees []repository.Tree, maxDistance, estateLength, estateWidth int) map[string]int {
	travelDistance := 1   // The drone starts with an initial elevation of 1m
	currentElevation := 1 // Start the drone at an elevation of 1 meter
	landingPoint := map[string]int{}

	// Start from plot (1,1)
	for y := 1; y <= estateWidth; y++ {
		// Move horizontally across the row (zigzag pattern)
		for x := 1; x <= estateLength; x++ {
			// Find the tree in this plot (or no tree if there's none)
			tree := findTreeAtPlot(trees, x, y)

			// If there's a tree, adjust the elevation accordingly
			if tree != nil {
				heightDifference := int(math.Abs(float64(tree.Height - currentElevation)))

				// Update the travel distance with the elevation difference
				travelDistance += heightDifference

				// Update the current elevation to match the tree's height + 1
				currentElevation = tree.Height + 1
			}

			// After each plot move, add 10 meters to the travel distance
			travelDistance += 10

			// Check if the drone has exceeded or reached the max distance
			if travelDistance >= maxDistance {
				// Found the landing point where the drone stops
				landingPoint = map[string]int{"x": x, "y": y}
				return landingPoint
			}
		}

		// After reaching the end of a row, zigzag back to the previous row
		estateLength, estateWidth = estateWidth, estateLength
	}

	// If we exit the loop and the max distance wasn't reached, return the last plot
	totalPlots := estateLength * estateWidth
	lastPlotX := (totalPlots - 1) % estateLength
	lastPlotY := (totalPlots - 1) / estateLength
	landingPoint = map[string]int{"x": lastPlotX + 1, "y": lastPlotY + 1}

	return landingPoint
}

// Helper function to find the tree at a specific plot coordinate
func findTreeAtPlot(trees []repository.Tree, x, y int) *repository.Tree {
	for _, tree := range trees {
		if tree.X == x && tree.Y == y {
			return &tree
		}
	}
	return nil // No tree at this plot
}
