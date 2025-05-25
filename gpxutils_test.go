package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/tkrajina/gpxgo/gpx"
)

const testFileDir = "testdata"
const numExpectedPoints = 1000 // Corresponds to numTargetPoints in gpxutils.go

// Helper function to compare two GPX points
func compareGPXPoints(t *testing.T, p1, p2 gpx.GPXPoint, msgAndArgs ...interface{}) {
	t.Helper()
	if math.Abs(p1.Latitude-p2.Latitude) > 1e-9 { // Use tolerance for float comparison
		t.Errorf("Latitude mismatch: expected %f, got %f. %s", p1.Latitude, p2.Latitude, fmt.Sprint(msgAndArgs...))
	}
	if math.Abs(p1.Longitude-p2.Longitude) > 1e-9 { // Use tolerance for float comparison
		t.Errorf("Longitude mismatch: expected %f, got %f. %s", p1.Longitude, p2.Longitude, fmt.Sprint(msgAndArgs...))
	}

	if p1.Elevation.Valid() != p2.Elevation.Valid() {
		t.Errorf("Elevation validity mismatch: expected %v, got %v. %s", p1.Elevation.Valid(), p2.Elevation.Valid(), fmt.Sprint(msgAndArgs...))
	}
	if p1.Elevation.Valid() { // Only compare values if valid (implicit that p2.Elevation.Valid() is also true due to above check)
		if math.Abs(p1.Elevation.Value()-p2.Elevation.Value()) > 0.001 { // Tolerance for float comparison
			t.Errorf("Elevation value mismatch: expected %f, got %f. %s", p1.Elevation.Value(), p2.Elevation.Value(), fmt.Sprint(msgAndArgs...))
		}
	}
}

func TestNormalizeGPX_SuccessfulNormalization(t *testing.T) {
	inputFile := filepath.Join(testFileDir, "sample.gpx")
	// expectedOutputFile is relative to the root where the test is run, not testFileDir
	expectedOutputFile := "normalized-" + filepath.Base(inputFile) // Created in repo root
	defer os.Remove(expectedOutputFile)

	err := normalizeGPX(inputFile, expectedOutputFile)
	if err != nil {
		t.Fatalf("normalizeGPX(%s, %s) failed: %v", inputFile, expectedOutputFile, err)
	}

	// Parse the output file
	normalizedGpxFile, err := gpx.ParseFile(expectedOutputFile)
	if err != nil {
		t.Fatalf("Error parsing normalized GPX file %s: %v", expectedOutputFile, err)
	}

	if len(normalizedGpxFile.Tracks) != 1 {
		t.Fatalf("Expected 1 track, got %d", len(normalizedGpxFile.Tracks))
	}
	if len(normalizedGpxFile.Tracks[0].Segments) != 1 {
		t.Fatalf("Expected 1 segment, got %d", len(normalizedGpxFile.Tracks[0].Segments))
	}
	if len(normalizedGpxFile.Tracks[0].Segments[0].Points) != numExpectedPoints {
		t.Fatalf("Expected %d points, got %d", numExpectedPoints, len(normalizedGpxFile.Tracks[0].Segments[0].Points))
	}

	// Verify first and last points
	originalGpxFile, err := gpx.ParseFile(inputFile)
	if err != nil {
		t.Fatalf("Error parsing original GPX file %s: %v", inputFile, err)
	}
	originalPoints := originalGpxFile.Tracks[0].Segments[0].Points
	normalizedPoints := normalizedGpxFile.Tracks[0].Segments[0].Points

	compareGPXPoints(t, originalPoints[0], normalizedPoints[0], "First point mismatch")
	compareGPXPoints(t, originalPoints[len(originalPoints)-1], normalizedPoints[numExpectedPoints-1], "Last point mismatch")

	// (Bonus) Basic equidistance check
	totalDistance := normalizedGpxFile.Tracks[0].Segments[0].Length2D()
	if totalDistance == 0 && len(normalizedPoints) > 1 { // Avoid division by zero if all points are same
		t.Logf("Total distance is 0, skipping equidistance check for distinct points.")
	} else if totalDistance > 0 {
		expectedInterval := totalDistance / float64(numExpectedPoints-1)
		
		testIntervals := [][2]int{{0, 1}, {numExpectedPoints / 2 -1, numExpectedPoints / 2}, {numExpectedPoints - 2, numExpectedPoints - 1}}

		for _, intervalIdx := range testIntervals {
			p1 := normalizedPoints[intervalIdx[0]]
			p2 := normalizedPoints[intervalIdx[1]]
			dist := p1.Distance2D(&p2)
			
			if expectedInterval > 1e-9 { 
				tolerance := 0.01 // 1% tolerance, as per instructions
				relativeDifference := math.Abs(dist - expectedInterval) / expectedInterval
				if relativeDifference > tolerance {
					t.Errorf("Equidistance check failed for points %d-%d: expected interval ~%.6f, got %.6f. Relative difference: %.6f > tolerance %.6f",
						intervalIdx[0], intervalIdx[1], expectedInterval, dist, relativeDifference, tolerance)
				}
			} else if dist > 1e-9 { 
                 t.Errorf("Equidistance check failed for points %d-%d: expected interval ~0 (<=1e-9), got %.6f (>1e-9)", 
				 	intervalIdx[0], intervalIdx[1], dist)
            }
		}
	}
}

func TestNormalizeGPX_LessThanTwoPoints(t *testing.T) {
	inputFile := filepath.Join(testFileDir, "one_point.gpx")
	outputFile := "normalized-one_point.gpx" 
	defer os.Remove(outputFile) 

	err := normalizeGPX(inputFile, outputFile)
	if err == nil {
		t.Errorf("Expected an error for GPX file with less than two points (%s), but got nil", inputFile)
	}
}

func TestNormalizeGPX_ZeroDistancePoints(t *testing.T) {
	inputFile := filepath.Join(testFileDir, "zero_dist.gpx")
	expectedOutputFile := "normalized-" + filepath.Base(inputFile) 
	defer os.Remove(expectedOutputFile)

	err := normalizeGPX(inputFile, expectedOutputFile)
	if err != nil {
		t.Fatalf("normalizeGPX(%s, %s) failed: %v", inputFile, expectedOutputFile, err)
	}

	normalizedGpxFile, err := gpx.ParseFile(expectedOutputFile)
	if err != nil {
		t.Fatalf("Error parsing normalized GPX file %s: %v", expectedOutputFile, err)
	}

	if len(normalizedGpxFile.Tracks) != 1 {
		t.Fatalf("Expected 1 track, got %d", len(normalizedGpxFile.Tracks))
	}
	if len(normalizedGpxFile.Tracks[0].Segments) != 1 {
		t.Fatalf("Expected 1 segment, got %d", len(normalizedGpxFile.Tracks[0].Segments))
	}
	normalizedPoints := normalizedGpxFile.Tracks[0].Segments[0].Points
	if len(normalizedPoints) != numExpectedPoints {
		t.Fatalf("Expected %d points, got %d", numExpectedPoints, len(normalizedPoints))
	}

	originalGpxFile, err := gpx.ParseFile(inputFile)
	if err != nil {
		t.Fatalf("Error parsing original GPX file %s: %v", inputFile, err)
	}
	firstOriginalPoint := originalGpxFile.Tracks[0].Segments[0].Points[0]

	for i, p := range normalizedPoints {
		compareGPXPoints(t, firstOriginalPoint, p, fmt.Sprintf("Point %d mismatch with first original point", i))
	}
}

func TestNormalizeGPX_NonExistentFile(t *testing.T) {
	inputFile := "non_existent_file.gpx"
	outputFile := "normalized-non_existent.gpx" 
	defer os.Remove(outputFile) 

	err := normalizeGPX(inputFile, outputFile)
	if err == nil {
		t.Errorf("Expected an error for non-existent input file (%s), but got nil", inputFile)
	}
}

// Added Test Function
func TestNormalizeGPX_LargeFile_3000Points(t *testing.T) {
	inputFile := filepath.Join(testFileDir, "large_sample.gpx")
	expectedOutputFile := "normalized-large_sample.gpx" // Output in repo root

	defer os.Remove(expectedOutputFile)

	err := normalizeGPX(inputFile, expectedOutputFile)
	if err != nil {
		t.Fatalf("normalizeGPX(%s, %s) failed: %v", inputFile, expectedOutputFile, err)
	}

	// Check if output file was created
	if _, errStat := os.Stat(expectedOutputFile); os.IsNotExist(errStat) {
		t.Fatalf("Expected output file %s was not created", expectedOutputFile)
	}

	normalizedGpxFile, err := gpx.ParseFile(expectedOutputFile)
	if err != nil {
		t.Fatalf("Failed to parse normalized GPX file %s: %v", expectedOutputFile, err)
	}

	if len(normalizedGpxFile.Tracks) != 1 {
		t.Fatalf("Expected 1 track in normalized file, got %d", len(normalizedGpxFile.Tracks))
	}
	normalizedTrack := normalizedGpxFile.Tracks[0]
	if len(normalizedTrack.Segments) != 1 {
		t.Fatalf("Expected 1 segment in normalized track, got %d", len(normalizedTrack.Segments))
	}
	normalizedSegment := normalizedTrack.Segments[0]
	normalizedPoints := normalizedSegment.Points

	if len(normalizedPoints) != numExpectedPoints { 
		t.Fatalf("Expected %d points in normalized segment, got %d", numExpectedPoints, len(normalizedPoints))
	}

	// First and Last point check
	originalGpxFile, err := gpx.ParseFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to parse original GPX file %s: %v", inputFile, err)
	}
	if len(originalGpxFile.Tracks) == 0 || len(originalGpxFile.Tracks[0].Segments) == 0 {
		t.Fatalf("Original GPX file %s has no tracks or segments", inputFile)
	}
	originalPoints := originalGpxFile.Tracks[0].Segments[0].Points
	if len(originalPoints) == 0 {
		t.Fatalf("Original GPX file %s has no points in the first segment", inputFile)
	}


	compareGPXPoints(t, normalizedPoints[0], originalPoints[0], "first point mismatch for large file")
	
	// The original file has 3000 points (0-2999). The normalized has 1000 (0-999).
	// Last point of normalized should match last point of original.
	compareGPXPoints(t, normalizedPoints[numExpectedPoints-1], originalPoints[len(originalPoints)-1], "last point mismatch for large file")
	
	// Equidistance check
	totalDistance := normalizedSegment.Length2D()
	if len(normalizedPoints) < 2 {
		t.Logf("Skipping equidistance check as there are less than 2 points in normalized output")
		return
	}
	expectedInterval := totalDistance / (float64(len(normalizedPoints) - 1))
	
	// Check a few intervals
	indicesToCheck := []int{0, len(normalizedPoints) / 2, len(normalizedPoints) - 2} 
	if len(normalizedPoints) <= 2 {
		indicesToCheck = []int{0} // Only check the first interval if 2 points
	}

	for _, idx := range indicesToCheck {
		// Ensure idx+1 is a valid index
		if idx+1 >= len(normalizedPoints) {
			continue 
		}
		p1 := normalizedPoints[idx]
		p2 := normalizedPoints[idx+1]
		dist := p1.Distance2D(&p2)

		if expectedInterval == 0 { // Handles cases like zero_dist.gpx or a track with all identical points
			if math.Abs(dist) > 1e-9 { // Allow very small tolerance for zero
				t.Errorf("Point %d to %d (large file): Expected interval distance ~0, got %f", idx, idx+1, dist)
			}
		} else {
			relativeError := math.Abs(dist-expectedInterval) / expectedInterval
			if relativeError > 0.01 { // 1% tolerance
				t.Errorf("Point %d to %d (large file): Expected interval distance %f, got %f (relative error: %.2f%%)", idx, idx+1, expectedInterval, dist, relativeError*100)
			}
		}
	}
}
