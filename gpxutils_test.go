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
	if p1.Latitude != p2.Latitude {
		t.Errorf("Latitude mismatch: expected %f, got %f. %s", p1.Latitude, p2.Latitude, fmt.Sprint(msgAndArgs...))
	}
	if p1.Longitude != p2.Longitude {
		t.Errorf("Longitude mismatch: expected %f, got %f. %s", p1.Longitude, p2.Longitude, fmt.Sprint(msgAndArgs...))
	}
	if p1.Elevation.NullFloat64.Valid != p2.Elevation.NullFloat64.Valid {
		t.Errorf("Elevation validity mismatch: expected %v, got %v. %s", p1.Elevation.NullFloat64.Valid, p2.Elevation.NullFloat64.Valid, fmt.Sprint(msgAndArgs...))
	}
	if p1.Elevation.NullFloat64.Valid && p2.Elevation.NullFloat64.Valid {
		if p1.Elevation.Value() != p2.Elevation.Value() {
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
			
			// Allow for some tolerance, especially if expectedInterval is very small
			// or for the very first/last segments which might have slight variations.
			// If expectedInterval is zero (e.g. only 1 unique point repeated), this check is skipped.
			// Using 1e-9 as a threshold for "effectively zero" distance or interval.
			if expectedInterval > 1e-9 { 
				tolerance := 0.01 // 1% tolerance, as per instructions
				relativeDifference := math.Abs(dist - expectedInterval) / expectedInterval
				if relativeDifference > tolerance {
					t.Errorf("Equidistance check failed for points %d-%d: expected interval ~%.6f, got %.6f. Relative difference: %.6f > tolerance %.6f",
						intervalIdx[0], intervalIdx[1], expectedInterval, dist, relativeDifference, tolerance)
				}
			// If expectedInterval is effectively zero, then dist should also be effectively zero.
			} else if dist > 1e-9 { 
                 t.Errorf("Equidistance check failed for points %d-%d: expected interval ~0 (<=1e-9), got %.6f (>1e-9)", 
				 	intervalIdx[0], intervalIdx[1], dist)
            }
		}
	}
}

func TestNormalizeGPX_LessThanTwoPoints(t *testing.T) {
	inputFile := filepath.Join(testFileDir, "one_point.gpx")
	outputFile := "normalized-one_point.gpx" // Will not be created if error occurs as expected
	defer os.Remove(outputFile) // Cleanup in case it is created

	err := normalizeGPX(inputFile, outputFile)
	if err == nil {
		t.Errorf("Expected an error for GPX file with less than two points (%s), but got nil", inputFile)
	}
}

func TestNormalizeGPX_ZeroDistancePoints(t *testing.T) {
	inputFile := filepath.Join(testFileDir, "zero_dist.gpx")
	expectedOutputFile := "normalized-" + filepath.Base(inputFile) // Created in repo root
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
	outputFile := "normalized-non_existent.gpx" // Will not be created
	defer os.Remove(outputFile) // Cleanup in case it is created

	err := normalizeGPX(inputFile, outputFile)
	if err == nil {
		t.Errorf("Expected an error for non-existent input file (%s), but got nil", inputFile)
	}
}
