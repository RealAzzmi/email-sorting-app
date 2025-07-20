# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AI-Powered Email Sorting and Management App with OAuth-based Google authentication, multi-account Gmail integration, custom AI-powered email categorization, automatic archiving, AI-generated summaries, and bulk email management with intelligent unsubscribe capabilities.

## Project Structure

- **frontend/**: Next.js 15 application with TypeScript, Tailwind CSS 4, and React 19
- **backend/**: Go application using Gin framework with pgx driver for PostgreSQL database

## Common Development Commands

### Frontend Development (run from `/frontend` directory)
```bash
# Start development server with Turbopack
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Run ESLint
npm run lint
```

### Backend Development (run from `/backend` directory)
```bash
# Start development server
go run cmd/server/main.go

# Or use the convenience script
./start.sh

# Build the application
go build -o email-sorting-app cmd/server/main.go

# Initialize database (requires PostgreSQL running)
psql -d email_sorting_app -f internal/adapters/database/migrations/schema.sql

# Set up environment (copy .env.example to .env and configure)
cp .env.example .env
```

## Technology Stack

### Frontend
- **Framework**: Next.js 15.4.1 with App Router
- **Language**: TypeScript 5
- **Styling**: Tailwind CSS 4 with PostCSS
- **Build Tool**: Turbopack (via `--turbopack` flag)
- **Runtime**: React 19.1.0
- **Linting**: ESLint with Next.js config

### Backend
- **Language**: Go 1.23
- **Web Framework**: Gin
- **Database Driver**: pgx v5 for PostgreSQL
- **ORM/Database Interaction**: Raw SQL with pgx
- **AI Integration**: Google Gemini API for email categorization
- **OAuth**: Google OAuth2 for Gmail authentication

### Configuration
- TypeScript path alias: `@/*` points to `./src/*`
- ESLint extends `next/core-web-vitals` and `next/typescript`
- Tailwind CSS with PostCSS integration

## Backend Architecture

The backend follows Clean Architecture with clear separation of concerns:

### Layer Structure
- **cmd/server/**: Application entry point and dependency injection
- **internal/domain/**: Core business entities and repository interfaces
- **internal/usecases/**: Business logic and use case implementations  
- **internal/adapters/**: External interfaces (HTTP handlers, database, Gmail API)

### Key Components
- **Repositories**: Data access layer with PostgreSQL using pgx driver
- **Use Cases**: Business logic layer that orchestrates between repositories and external services
- **Handlers**: HTTP request/response handling via Gin framework
- **Gmail Service**: Integration with Google Gmail API for email operations
- **AI Service**: Google Gemini integration for email categorization and summaries

### Database Schema
- **accounts**: OAuth-authenticated Gmail accounts with tokens
- **emails**: Email messages with AI summaries and metadata
- **categories**: User-defined categories for email organization
- **email_categories**: Many-to-many relationship between emails and categories

### Environment Setup
Required environment variables (see `.env.example`):
- `DATABASE_URL`: PostgreSQL connection string
- `GOOGLE_CLIENT_ID` & `GOOGLE_CLIENT_SECRET`: Google OAuth credentials
- `GEMINI_API_KEY`: Google Gemini API key for AI categorization
- `REDIRECT_URL`: OAuth callback URL
- `PORT`: Server port (default 8080)

## Development Guidelines

- Use concise code
- Follow clean architecture principles
- Minimize testing overhead
- Do minimal logging
- Use best and modern practices according to official documentation when possible

## Design Preferences

- Use 3 colors: black, white, and light gray
- Prefer minimalist design approach