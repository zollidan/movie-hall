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
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload" // Automatically load .env file

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

// writeError отправляет JSON ответ с ошибкой и указанным HTTP статус кодом
func (d DB) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// writeSuccess отправляет JSON ответ с сообщением об успешной операции
func (d DB) writeSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: message})
}

// getLibraryPath возвращает путь к библиотеке фильмов из настроек
func (d DB) getLibraryPath() (string, error) {
	var settings Settings
	err := d.db.First(&settings).Error
	if err != nil {
		return "", err
	}
	return settings.LibPath, nil
}

// scanLibraryDirectory сканирует директорию с фильмами и добавляет новые фильмы в базу данных
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
			var count int64
			err := d.db.Model(&Movies{}).Where("title = ?", e.Name()).Count(&count).Error
			if err != nil {
				return fmt.Errorf("failed to check existing movie: %w", err)
			}

			if count == 0 {
				parsed := parseMovieTitle(e.Name())
				movie := Movies{
					Title: parsed.Title,
					Year:  parsed.Year,
				}

				// Try to fetch additional info from OMDB
				if omdbInfo, err := fetchMovieInfo(parsed.Title, parsed.Year); err == nil {
					movie.Title = omdbInfo.Title // Use official title
					if year, err := strconv.Atoi(omdbInfo.Year); err == nil {
						movie.Year = year
					}
					movie.Cover = omdbInfo.Poster
				} else {
					log.Printf("Failed to fetch OMDB info for %s: %v", parsed.Title, err)
				}

				err = d.db.Create(&movie).Error
				if err != nil {
					return fmt.Errorf("failed to create movie record: %w", err)
				}
				log.Printf("Added movie: %s (%d)", movie.Title, movie.Year)
			}
		}
	}

	return nil
}

// showLibrary возвращает список всех фильмов из библиотеки. При первом вызове автоматически сканирует директорию
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

// showSettings возвращает все настройки приложения в формате JSON
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

// setSettings создает или обновляет настройки приложения, включая путь к библиотеке фильмов
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

// rescanLibrary принудительно пересканирует библиотеку фильмов и добавляет новые файлы в базу данных
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

// refreshMovieInfo refreshes movie information from OMDB API for a specific movie
func (d DB) refreshMovieInfo(w http.ResponseWriter, r *http.Request) {
	var movie Movies
	movieID := chi.URLParam(r, "id")

	if err := d.db.First(&movie, movieID).Error; err != nil {
		d.writeError(w, http.StatusNotFound, "Movie not found")
		return
	}

	if omdbInfo, err := fetchMovieInfo(movie.Title, movie.Year); err == nil {
		movie.Title = omdbInfo.Title
		if year, err := strconv.Atoi(omdbInfo.Year); err == nil {
			movie.Year = year
		}
		movie.Cover = omdbInfo.Poster

		if err := d.db.Save(&movie).Error; err != nil {
			d.writeError(w, http.StatusInternalServerError, "Failed to update movie")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movie)
	} else {
		d.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch OMDB info: %v", err))
	}
}

// main инициализирует базу данных, настраивает HTTP маршруты и запускает веб-сервер
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
	r.Use(cors.Handler(cors.Options{

		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Route("/api", func(r chi.Router) {
		r.Get("/library", app.showLibrary)
		r.Post("/library/rescan", app.rescanLibrary)
		r.Get("/settings", app.showSettings)
		r.Post("/settings", app.setSettings)

		r.Post("/movies/{id}/refresh", app.refreshMovieInfo)
	})

	log.Printf("Starting server on http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
