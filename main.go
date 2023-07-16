package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Line represents a line with start and end points.
type Line struct {
	ID    string `json:"id"`
	Start Point  `json:"start"`
	End   Point  `json:"end"`
}

// Point represents a point in GeoJSON.
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// IntersectionResult represents the result of the intersection check.
type IntersectionResult struct {
	LineID       string `json:"lineId"`
	Intersection Point  `json:"intersection"`
}

// IntersectingLinesResponse represents the response for intersecting lines.
type IntersectingLinesResponse struct {
	Intersections []IntersectionResult `json:"intersections"`
}

// LineString represents a GeoJSON linestring.
type LineString struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

// LineData represents the data for a line.
type LineData struct {
	Line LineString `json:"line"`
}

func main() {
	http.HandleFunc("/intersect", IntersectHandler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

// IntersectHandler handles the intersection check API endpoint.
func IntersectHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	authHeader := r.Header.Get("Authorization")
	if !isValidAuthHeader(authHeader) {
		http.Error(w, "Invalid authentication header", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var linestring []Point
	if err := json.Unmarshal(body, &linestring); err != nil {
		http.Error(w, "Invalid linestring format", http.StatusBadRequest)
		return
	}

	linesData := []LineData{
		// Add the provided lines data here...
		LineData{
			Line: LineString{
				Type: "LineString",
				Coordinates: [][]float64{
					{-74.0386542, 40.7302174},
					{-74.038756, 40.7295611},
				},
			},
		},
		LineData{
			Line: LineString{
				Type: "LineString",
				Coordinates: [][]float64{
					{-74.061602, 40.705933},
					{-74.06214, 40.706563},
				},
			},
		},
	}

	lines := convertToLines(linesData)

	intersections := findIntersectingLines(linestring, lines)

	response := IntersectingLinesResponse{Intersections: intersections}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error preparing response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// isValidAuthHeader checks if the Authorization header is valid for authentication.
func isValidAuthHeader(header string) bool {

	const validToken = "a45G7hJ3lO1pR9tU2xY8zAbC6eF0"

	// Check the value of the header and the valid token for debugging.
	fmt.Println("Header Value:", header)
	//fmt.Println("Valid Token:", "Bearer "+validToken)

	// Check if the header value matches the valid token.
	return header == "Bearer "+validToken
}

func convertToLines(linesData []LineData) []Line {
	lines := make([]Line, len(linesData))
	for i, lineData := range linesData {
		lines[i] = Line{
			ID:    fmt.Sprintf("L%02d", i+1),
			Start: Point{Lat: lineData.Line.Coordinates[0][1], Lng: lineData.Line.Coordinates[0][0]},
			End:   Point{Lat: lineData.Line.Coordinates[1][1], Lng: lineData.Line.Coordinates[1][0]},
		}
	}
	return lines
}

func findIntersectingLines(linestring []Point, lines []Line) []IntersectionResult {
	intersections := []IntersectionResult{}
	for _, line := range lines {
		for i := 1; i < len(linestring); i++ {
			intersection := findIntersection(linestring[i-1], linestring[i], line.Start, line.End)
			if intersection != nil {
				intersections = append(intersections, IntersectionResult{
					LineID:       line.ID,
					Intersection: *intersection,
				})
			}
		}
	}
	return intersections
}

func findIntersection(p1, p2, q1, q2 Point) *Point {
	p1p2 := Point{Lat: p2.Lat - p1.Lat, Lng: p2.Lng - p1.Lng}
	q1q2 := Point{Lat: q2.Lat - q1.Lat, Lng: q2.Lng - q1.Lng}

	det := p1p2.Lat*q1q2.Lng - p1p2.Lng*q1q2.Lat

	if det == 0 {
		return nil
	}

	t := (q1.Lat-p1.Lat)*q1q2.Lng - (q1.Lng-p1.Lng)*q1q2.Lat
	t /= det

	u := (q1.Lat-p1.Lat)*p1p2.Lng - (q1.Lng-p1.Lng)*p1p2.Lat
	u /= det

	if t >= 0 && t <= 1 && u >= 0 && u <= 1 {
		intersection := Point{
			Lat: p1.Lat + t*p1p2.Lat,
			Lng: p1.Lng + t*p1p2.Lng,
		}
		return &intersection
	}

	return nil
}
