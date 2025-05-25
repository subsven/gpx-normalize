package main

import (
	"fmt"
	"os"
	"path/filepath"
	"math" // Added for math operations

	"github.com/tkrajina/gpxgo/gpx"
)

const numTargetPoints = 1000

func normalizeGPX(inputFile string, outputFile string) error {
	// Read the GPX file
	gpxData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading GPX file %s: %w", inputFile, err)
	}
	gpxFile, err := gpx.ParseBytes(gpxData)
	if err != nil {
		return fmt.Errorf("error parsing GPX file %s: %w", inputFile, err)
	}

	// Get the first track and segment
	if len(gpxFile.Tracks) == 0 {
		return fmt.Errorf("no tracks found in GPX file %s", inputFile)
	}
	track := &gpxFile.Tracks[0]

	if len(track.Segments) == 0 {
		return fmt.Errorf("no segments found in track of GPX file %s", inputFile)
	}
	segment := &track.Segments[0]
	sourcePoints := segment.Points

	// Check if the segment has at least 2 points
	if len(sourcePoints) < 2 {
		return fmt.Errorf("not enough points in GPX file %s (found %d, need at least 2)", inputFile, len(sourcePoints))
	}

	// Create a new GPX object
	newGpx := &gpx.GPX{}
	newGpx.Creator = "gpx-normalizer"
	newGpx.Version = gpxFile.Version
	newGpx.Name = gpxFile.Name
	newGpx.Description = gpxFile.Description
	newGpx.AuthorName = gpxFile.AuthorName
	newGpx.CopyrightAuthor = gpxFile.CopyrightAuthor
	newGpx.CopyrightYear = gpxFile.CopyrightYear
	newGpx.CopyrightLicense = gpxFile.CopyrightLicense
	newGpx.Link = gpxFile.Link
	newGpx.LinkText = gpxFile.LinkText
	newGpx.Time = gpxFile.Time
	newGpx.Keywords = gpxFile.Keywords
	newGpx.Bounds = gpxFile.Bounds
	newGpx.Extensions = gpxFile.Extensions


	// Create a new GPXTrack and add it to the new GPX object
	newTrack := gpx.GPXTrack{}
	newGpx.Tracks = append(newGpx.Tracks, newTrack)

	// Create a new GPXTrackSegment and add it to the new track
	newSegment := gpx.GPXTrackSegment{}
	newGpx.Tracks[0].Segments = append(newGpx.Tracks[0].Segments, newSegment)
	newSegmentPoints := &newGpx.Tracks[0].Segments[0].Points // Pointer to the new points slice

	totalDistance := segment.Length2D()

	// Handle zero total distance
	if totalDistance == 0 {
		if len(sourcePoints) > 0 {
			firstPoint := sourcePoints[0]
			for i := 0; i < numTargetPoints; i++ {
				*newSegmentPoints = append(*newSegmentPoints, firstPoint)
			}
		}
		// Proceed to write the file and return (handled later)
	} else {
		intervalDistance := totalDistance / float64(numTargetPoints-1)
		cumulativeDistance := 0.0
		originalPointIndex := 0

		for i := 0; i < numTargetPoints; i++ {
			var newPoint gpx.GPXPoint

			if i == 0 {
				newPoint = sourcePoints[0]
			} else if i == numTargetPoints-1 {
				newPoint = sourcePoints[len(sourcePoints)-1]
			} else {
				targetDistForCurrentPoint := float64(i) * intervalDistance

				// Advance originalPointIndex:
				// Loop while originalPointIndex is not the second to last point AND
				// the next segment's end (cumulativeDistance + distance to next point) is still less than our target.
				for originalPointIndex < len(sourcePoints)-2 && // Ensures sourcePoints[originalPointIndex+1] is valid
					cumulativeDistance+sourcePoints[originalPointIndex].Distance2D(&sourcePoints[originalPointIndex+1]) < targetDistForCurrentPoint {
					cumulativeDistance += sourcePoints[originalPointIndex].Distance2D(&sourcePoints[originalPointIndex+1])
					originalPointIndex++
				}

				p1 := sourcePoints[originalPointIndex]
				p2 := sourcePoints[originalPointIndex+1] // Safe because originalPointIndex <= len(sourcePoints)-2

				distToP1 := cumulativeDistance // Cumulative distance *to the start of the current segment (p1)*
				distP1P2 := p1.Distance2D(&p2)

				ratio := 0.0
				if distP1P2 > 0 {
					// ratio is how far along the segment (p1 to p2) our targetDistForCurrentPoint falls
					ratio = (targetDistForCurrentPoint - distToP1) / distP1P2
				}
				// Clamp ratio to [0, 1] to handle floating point inaccuracies or edge cases
				if ratio < 0 { ratio = 0 }
				if ratio > 1 { ratio = 1 }

				newLat := p1.Latitude + ratio*(p2.Latitude-p1.Latitude)
				newLon := p1.Longitude + ratio*(p2.Longitude-p1.Longitude)
				
				newEle := 0.0
				p1EleValid := p1.Elevation.NullFloat64.Valid
				p2EleValid := p2.Elevation.NullFloat64.Valid
				elevationInterpolated := false

				if p1EleValid && p2EleValid {
					newEle = p1.Elevation.Value() + ratio*(p2.Elevation.Value()-p1.Elevation.Value())
					elevationInterpolated = true
				} else if p1EleValid {
					newEle = p1.Elevation.Value()
					elevationInterpolated = true
				} else if p2EleValid {
					newEle = p2.Elevation.Value()
					elevationInterpolated = true
				}

				newPoint = gpx.GPXPoint{Latitude: newLat, Longitude: newLon, Timestamp: p1.Timestamp} // Use p1's timestamp

				if math.IsNaN(newPoint.Latitude) || math.IsNaN(newPoint.Longitude) {
					// Fallback if interpolation results in NaN (e.g., p1 and p2 are identical)
					newPoint.Latitude = p1.Latitude
					newPoint.Longitude = p1.Longitude
				}

				if elevationInterpolated {
					newPoint.Elevation = *gpx.NewNullableFloat64(newEle)
				}
			}
			*newSegmentPoints = append(*newSegmentPoints, newPoint)
		}
	}

	// Ensure newSegment.Points has numTargetPoints (mostly a safeguard).
	// This padding/truncating should ideally not be needed if the main loop is correct.
	// For totalDistance == 0, it's explicitly handled to fill all points.
	// This is mostly a safeguard; the logic above should handle it for totalDistance > 0.
	// For totalDistance == 0, it's explicitly handled.
	if len(*newSegmentPoints) < numTargetPoints && len(sourcePoints) > 0 {
		lastPt := (*newSegmentPoints)[len(*newSegmentPoints)-1]
		for len(*newSegmentPoints) < numTargetPoints {
			*newSegmentPoints = append(*newSegmentPoints, lastPt)
		}
	} else if len(*newSegmentPoints) > numTargetPoints { // Truncate if somehow we overshot
		*newSegmentPoints = (*newSegmentPoints)[:numTargetPoints]
	}


	// Convert the new GPX object to XML bytes
	xmlBytes, err := newGpx.ToXml(gpx.ToXmlParams{Version: "1.1", Indent: true})
	if err != nil {
		return fmt.Errorf("error converting GPX to XML for %s: %w", outputFile, err)
	}

	// Write the XML bytes to the output file
	err = os.WriteFile(outputFile, xmlBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing normalized GPX file %s: %w", outputFile, err)
	}

	return nil
}
