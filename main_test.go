package main

import (
	"image"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestParseWidthHeight tests the parseWidthHeight function.
func TestParseWidthHeight(t *testing.T) {
	tests := []struct {
		query          string
		expectedWidth  string
		expectedHeight string
	}{
		{"400x300", "400", "300"},
		{"400x", "400", ""},
		{"x300", "", "300"},
		{"", "", ""},
	}

	for _, test := range tests {
		width, height := parseWidthHeight(test.query)
		if width != test.expectedWidth || height != test.expectedHeight {
			t.Errorf("parseWidthHeight(%q) = %q, %q; want %q, %q", test.query, width, height, test.expectedWidth, test.expectedHeight)
		}
	}
}

// TestFindImagePath tests the findImagePath function.
func TestFindImagePath(t *testing.T) {
	tempDir := t.TempDir()

	// Create dummy image files
	imageFiles := []string{"test.jpg", "test.png", "test.webp"}
	for _, file := range imageFiles {
		f, err := os.Create(filepath.Join(tempDir, file))
		if err != nil {
			t.Fatalf("Failed to create temp image file: %v", err)
		}
		f.Close()
	}

	tests := []struct {
		baseDir   string
		basePath  string
		expect    string
		shouldErr bool
	}{
		{tempDir, "test", filepath.Join(tempDir, "test.jpg"), false},
		{tempDir, "nonexistent", "", true},
		{"invalid_base_dir", "test", "", true},
	}

	for _, test := range tests {
		result, err := findImagePath(test.baseDir, test.basePath)
		if test.shouldErr && err == nil {
			t.Errorf("Expected an error but got none for baseDir: %q, basePath: %q", test.baseDir, test.basePath)
		} else if !test.shouldErr && err != nil {
			t.Errorf("Did not expect an error but got one for baseDir: %q, basePath: %q", test.baseDir, test.basePath)
		} else if result != test.expect {
			t.Errorf("findImagePath(%q, %q) = %q; want %q", test.baseDir, test.basePath, result, test.expect)
		}
	}
}

// TestServeImage tests the serveImage function.
func TestServeImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	tests := []struct {
		format    string
		expectCT  string
		shouldErr bool
	}{
		{"jpeg", "image/jpeg", false},
		{"png", "image/png", false},
		{"webp", "image/webp", false},
		{"unsupported", "", true},
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()
		serveImage(rr, img, test.format)

		if test.shouldErr {
			if rr.Code != http.StatusBadRequest {
				t.Errorf("serveImage() = status %d; want %d", rr.Code, http.StatusBadRequest)
			}
		} else {
			if ct := rr.Header().Get("Content-Type"); ct != test.expectCT {
				t.Errorf("serveImage() = Content-Type %q; want %q", ct, test.expectCT)
			}
		}
	}
}

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
