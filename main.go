package main

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
	"gopkg.in/ini.v1"
)

type RouteConfig struct {
	Route string
	Dir   string
}

var (
	routes  []RouteConfig
	port    string
	baseDir string
)

func main() {
	if err := loadConfig("config.ini"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	http.HandleFunc("/health", healthCheckHandler)

	for _, route := range routes {
		log.Printf("Setting up handler for route: %s with directory: %s", route.Route, route.Dir)
		http.HandleFunc(route.Route, makeResizeHandler(route.Route, route.Dir))
	}

	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("Server started on :%s", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func loadConfig(filePath string) error {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return err
	}

	port = cfg.Section("server").Key("port").String()
	baseDir = cfg.Section("server").Key("baseDir").String()

	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection || section.Name() == "server" {
			continue
		}
		route := section.Key("route").String()
		dir := section.Key("dir").String()
		if route != "" && dir != "" {
			routes = append(routes, RouteConfig{Route: route, Dir: dir})
		}
	}

	return nil
}

func parseWidthHeight(query string) (string, string) {
	parts := strings.SplitN(query, "x", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func makeResizeHandler(baseRoute, baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, baseRoute)
		log.Printf("Requested path: %s\n", path)

		requestedExt := filepath.Ext(path)
		basePath := strings.TrimSuffix(path, requestedExt)
		widthStr, heightStr := parseWidthHeight(r.URL.RawQuery)
		if widthStr == "" {
			widthStr = r.URL.Query().Get("width")
		}
		if heightStr == "" {
			heightStr = r.URL.Query().Get("height")
		}
		if widthStr == "" && heightStr == "" {
			serveOriginalImage(w, baseDir, basePath, requestedExt)
			return
		}

		width, height, err := parseDimensions(widthStr, heightStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		format := strings.ToLower(strings.TrimPrefix(requestedExt, "."))
		imgPath, err := findImagePath(baseDir, basePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		img, err := decodeImage(imgPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Error decoding image %s: %v", imgPath, err)
			return
		}

		newImg := resizeImage(img, width, height)
		serveImage(w, newImg, format)
	}
}

func parseDimensions(widthStr, heightStr string) (int, int, error) {
	width, err := strconv.Atoi(widthStr)
	if err != nil {
		return 0, 0, err
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

func serveOriginalImage(w http.ResponseWriter, baseDir, basePath, requestedExt string) {
	format := strings.ToLower(strings.TrimPrefix(requestedExt, "."))
	imgPath, err := findImagePath(baseDir, basePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	img, err := decodeImage(imgPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error decoding image %s: %v", imgPath, err)
		return
	}

	serveImage(w, img, format)
}

func decodeImage(imgPath string) (image.Image, error) {
	imgPath = filepath.Clean(imgPath)

	file, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func serveImage(w http.ResponseWriter, img image.Image, format string) {
	switch format {
	case "jpeg", "jpg":
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, img, &jpeg.Options{Quality: 85})
	case "png":
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	case "webp":
		w.Header().Set("Content-Type", "image/webp")
		webp.Encode(w, img, &webp.Options{Quality: 85})
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
	}
}

func findImagePath(baseDir, basePath string) (string, error) {
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp"} {
		potentialPath := filepath.Join(baseDir, basePath+ext)
		cleanPath := filepath.Clean(potentialPath)

		if !strings.HasPrefix(cleanPath, filepath.Clean(baseDir)) {
			return "", errors.New("invalid path")
		}

		if _, err := os.Stat(cleanPath); err == nil {
			return cleanPath, nil
		}
	}
	return "", errors.New("file not found")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func resizeImage(img image.Image, width, height int) image.Image {
	if width == 0 {
		return resize.Resize(0, uint(height), img, resize.Lanczos3)
	}
	if height == 0 {
		return resize.Resize(uint(width), 0, img, resize.Lanczos3)
	}
	return resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
}
