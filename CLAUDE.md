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

## Technology Stack

### Frontend
- **Framework**: Next.js 15.4.1 with App Router
- **Language**: TypeScript 5
- **Styling**: Tailwind CSS 4 with PostCSS
- **Build Tool**: Turbopack (via `--turbopack` flag)
- **Runtime**: React 19.1.0
- **Linting**: ESLint with Next.js config

### Backend
- **Language**: Go
- **Web Framework**: Gin
- **Database Driver**: pgx for PostgreSQL
- **ORM/Database Interaction**: Raw SQL with pgx

### Configuration
- TypeScript path alias: `@/*` points to `./src/*`
- ESLint extends `next/core-web-vitals` and `next/typescript`
- Tailwind CSS with PostCSS integration

## Development Guidelines

- Use concise code
- Follow clean architecture principles
- Minimize testing overhead
- Do minimal logging
- Use best and modern practices according to official documentation when possible

## Design Preferences

- Use 3 colors: black, white, and light gray
- Prefer minimalist design approach