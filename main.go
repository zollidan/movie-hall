package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var addr = "localhost:8080"
var videoFormats = []string{".mkv", ".avi", ".mp4"}
var OMDBUrl = "http://www.omdbapi.com/"

type DB struct {
	db  *gorm.DB
	ctx context.Context
}

type Settings struct {
	gorm.Model
	LibPath string
}

type Movies struct {
	gorm.Model
	Title string
	Year  int
	Cover string
}

type SettingsRequest struct {
	LibPath string `json:"libPath"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func (d DB) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (d DB) writeSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: message})
}

func (d DB) getLibraryPath() (string, error) {
	var settings Settings
	err := d.db.First(&settings).Error
	if err != nil {
		return "", err
	}
	return settings.LibPath, nil
}

func (d DB) scanLibraryDirectory(libPath string) error {
	entries, err := os.ReadDir(libPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("directory is empty")
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(e.Name()))
		if slices.Contains(videoFormats, ext) {
			// Проверяем, не существует ли уже такой фильм
			var count int64
			err := d.db.Model(&Movies{}).Where("title = ?", e.Name()).Count(&count).Error
			if err != nil {
				return fmt.Errorf("failed to check existing movie: %w", err)
			}

			if count == 0 {
				movie := Movies{Title: e.Name()}
				err = d.db.Create(&movie).Error
				if err != nil {
					return fmt.Errorf("failed to create movie record: %w", err)
				}
				log.Printf("Added movie: %s", e.Name())
			}
		}
	}

	return nil
}

func (d DB) showLibrary(w http.ResponseWriter, r *http.Request) {
	var moviesCount int64
	err := d.db.Model(&Movies{}).Count(&moviesCount).Error
	if err != nil {
		d.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var settingsCount int64
	err = d.db.Model(&Settings{}).Count(&settingsCount).Error
	if err != nil {
		d.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if settingsCount == 0 {
		d.writeError(w, http.StatusBadRequest, "Setup app first")
		return
	}

	if moviesCount == 0 {
		libPath, err := d.getLibraryPath()
		if err != nil {
			d.writeError(w, http.StatusInternalServerError, "Failed to get library path")
			return
		}

		err = d.scanLibraryDirectory(libPath)
		if err != nil {
			d.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	var movies []Movies
	err = d.db.Find(&movies).Error
	if err != nil {
		d.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func (d DB) showSettings(w http.ResponseWriter, r *http.Request) {
	var settings []Settings
	err := d.db.Find(&settings).Error
	if err != nil {
		d.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (d DB) setSettings(w http.ResponseWriter, r *http.Request) {
	var requestBody SettingsRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		d.writeError(w, http.StatusBadRequest, "Error decoding request body: "+err.Error())
		return
	}

	if requestBody.LibPath == "" {
		d.writeError(w, http.StatusBadRequest, "LibPath cannot be empty")
		return
	}

	if _, err := os.Stat(requestBody.LibPath); os.IsNotExist(err) {
		d.writeError(w, http.StatusBadRequest, "Directory does not exist")
		return
	}

	var existingSettings Settings
	err = d.db.First(&existingSettings).Error

	if err == gorm.ErrRecordNotFound {
		settings := Settings{
			LibPath: requestBody.LibPath,
		}
		err = d.db.Create(&settings).Error
		if err != nil {
			d.writeError(w, http.StatusInternalServerError, "Error creating settings")
			return
		}
	} else if err != nil {
		d.writeError(w, http.StatusInternalServerError, "Error checking existing settings")
		return
	} else {
		existingSettings.LibPath = requestBody.LibPath
		err = d.db.Save(&existingSettings).Error
		if err != nil {
			d.writeError(w, http.StatusInternalServerError, "Error updating settings")
			return
		}
	}

	d.writeSuccess(w, "Settings saved successfully")
}

func (d DB) rescanLibrary(w http.ResponseWriter, r *http.Request) {
	libPath, err := d.getLibraryPath()
	if err != nil {
		d.writeError(w, http.StatusBadRequest, "No library path configured")
		return
	}

	err = d.scanLibraryDirectory(libPath)
	if err != nil {
		d.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	d.writeSuccess(w, "Library rescanned successfully")
}

func main() {
	db, err := gorm.Open(sqlite.Open("movs.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	ctx := context.Background()
	app := DB{db: db, ctx: ctx}

	err = db.AutoMigrate(&Settings{}, &Movies{})
	if err != nil {
		panic("failed to migrate database")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/api", func(r chi.Router) {
		r.Get("/library", app.showLibrary)
		r.Post("/library/rescan", app.rescanLibrary)
		r.Get("/settings", app.showSettings)
		r.Post("/settings", app.setSettings)
	})

	log.Printf("Starting server on http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
