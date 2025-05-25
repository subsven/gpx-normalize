# GPX Normalizer

A Go-based CLI application that reads GPX files and normalizes each track to contain exactly 1000 equidistant points.

## Features

- Processes multiple GPX files specified as command-line arguments.
- Normalizes each GPX track to 1000 equidistant points.
- Assumes constant speed between the start and end time of the track (timestamps are not deeply analyzed but are preserved from the source for interpolated points where possible).
- Saves processed files with the prefix "normalized-" (e.g., `normalized-myroute.gpx`).
- Uses Go routines for concurrent processing of multiple files.

## Prerequisites

- Go (version 1.18 or later recommended)

## Build Instructions

To build the application:
```bash
go build
```
This will create an executable named `gpx-normalizer` (or `gpx-normalizer.exe` on Windows) in the current directory.

## Usage

Run the application from the command line, providing one or more GPX files as arguments:

```bash
./gpx-normalizer <file1.gpx> [file2.gpx] [file3.gpx] ...
```

Example:
```bash
./gpx-normalizer myride.gpx another_trip.gpx
```

The application will output:
- `normalized-myride.gpx`
- `normalized-another_trip.gpx`

Log messages indicating the status of each file will be printed to the console.

## GPX Library Used

This project uses the [`tkrajina/gpxgo`](https://github.com/tkrajina/gpxgo) library for parsing and manipulating GPX files.

## Development

### Running Tests
To run the automated tests:
```bash
go test
```
