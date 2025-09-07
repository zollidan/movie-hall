# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Movie Hall is a full-stack movie library management application with a Go backend and React TypeScript frontend. The backend scans local movie directories and maintains a database of movie files, while the frontend provides a web interface to browse the movie collection.

## Architecture

### Backend (`backend/`)
- **Language**: Go 1.24.1
- **Framework**: Chi router with CORS middleware
- **Database**: SQLite with GORM
- **Server**: Runs on `localhost:8080`
- **Key files**:
  - `main.go`: HTTP server, API routes, database operations, movie scanning
  - `parser.go`: Movie title parsing logic with regex patterns
  - `movs.db`: SQLite database file

### Frontend (`frontend/`)
- **Framework**: React 19 + TypeScript + Vite
- **Styling**: Tailwind CSS v4 with custom components
- **UI Components**: Radix UI primitives
- **Features**: Dark/light theme toggle, responsive movie grid

## Development Commands

### Frontend
```bash
cd frontend
npm run dev        # Start development server
npm run build      # Build for production (runs tsc -b && vite build)
npm run lint       # Run ESLint
npm run preview    # Preview production build
```

### Backend
```bash
cd backend
go run .          # Run development server
go build          # Build binary
```

## API Endpoints

All API endpoints are prefixed with `/api`:

- `GET /api/library` - Get all movies (auto-scans on first request)
- `POST /api/library/rescan` - Force rescan movie directory
- `GET /api/settings` - Get application settings
- `POST /api/settings` - Set library path and other settings

## Database Models

- **Settings**: Stores library path configuration
- **Movies**: Stores movie title, year, and cover information

## Movie File Processing

The backend automatically scans configured directories for video files (`.mkv`, `.avi`, `.mp4`) and parses movie information using regex patterns in `parser.go`. Supported filename patterns:
- `Movie.Name.2023.quality.info.mkv`
- `Movie Name [2023 quality info]`  
- `Movie Name (2023)`

## Component Structure

Frontend uses shadcn/ui component library with:
- Theme provider for dark/light mode
- Movie grid with card components
- Skeleton loading states
- Custom UI components in `src/components/ui/`

## Key Dependencies

**Backend**: Chi router, GORM, SQLite driver
**Frontend**: React, Vite, Tailwind CSS, Radix UI, Lucide icons