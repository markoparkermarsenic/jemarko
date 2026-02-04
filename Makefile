.PHONY: help install setup start stop restart clean db-reset db-status logs

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Available commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo ""

## install: Install all dependencies (Vercel CLI, Supabase CLI, and npm packages)
install:
	@echo "📦 Installing dependencies..."
	@echo ""
	@echo "Installing Vercel CLI..."
	npm install -g vercel
	@echo ""
	@echo "Installing Supabase CLI..."
	@if command -v brew > /dev/null; then \
		brew install supabase/tap/supabase; \
	else \
		echo "⚠️  Homebrew not found. Please install Supabase CLI manually:"; \
		echo "   https://supabase.com/docs/guides/cli"; \
	fi
	@echo ""
	@echo "Installing frontend dependencies..."
	cd src && npm install
	@echo ""
	@echo "✅ All dependencies installed!"

## setup: Initial setup - initialize Supabase
setup:
	@echo "🔧 Setting up local environment..."
	@if [ ! -f "supabase/.gitignore" ]; then \
		echo "Initializing Supabase..."; \
		supabase init; \
	else \
		echo "Supabase already initialized"; \
	fi
	@echo ""
	@echo "Creating .env file from example..."
	@if [ ! -f "api/.env" ]; then \
		cp api/.env.example api/.env; \
		echo "✅ Created api/.env from example"; \
		echo "⚠️  Please edit api/.env and add your API keys"; \
	else \
		echo "api/.env already exists"; \
	fi
	@echo ""
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit api/.env and add your API keys (optional for local dev)"
	@echo "2. Run 'make start' to start the development servers"

## start: Start local development environment (Supabase + Vercel Dev)
start:
	@echo "🚀 Starting development environment..."
	@echo ""
	@echo "Starting Supabase..."
	@supabase start
	@echo ""
	@echo "✅ Supabase started!"
	@echo ""
	@echo "Supabase credentials:"
	@supabase status
	@echo ""
	@echo "importing csv!"
	@echo ""
	@make import-csv
	@echo ""
	@echo "Starting Vercel dev server (frontend + API functions)..."
	@echo "Press Ctrl+C to stop"
	@echo ""
	@vercel dev 

## start-frontend-only: Start only the frontend (no backend APIs)
start-frontend-only:
	@echo "🚀 Starting frontend only..."
	@echo "⚠️  Backend APIs will not be available"
	@echo ""
	cd src && npm run dev

## stop: Stop all local services
stop:
	@echo "🛑 Stopping local services..."
	@supabase stop
	@echo "✅ All services stopped!"

## restart: Restart all local services
restart: stop start

## db-reset: Reset Supabase database to initial state
db-reset:
	@echo "🔄 Resetting database..."
	@supabase db reset
	@echo "✅ Database reset complete!"

## db-status: Show Supabase status and credentials
db-status:
	@echo "📊 Supabase Status:"
	@echo ""
	@supabase status

## logs: Show Supabase logs
logs:
	@echo "📜 Supabase logs (press Ctrl+C to exit):"
	@supabase logs

## clean: Clean up all local data and stop services
clean:
	@echo "🧹 Cleaning up..."
	@supabase stop
	@echo "✅ Cleanup complete!"

## test-api: Test the API endpoints
test-api:
	@echo "🧪 Testing API endpoints..."
	@echo ""
	@echo "Health check:"
	@curl -s http://localhost:3000/api/health | jq .
	@echo ""
	@echo "Verify name test:"
	@curl -s -X POST http://localhost:3000/api/verify-name \
		-H "Content-Type: application/json" \
		-d '{"name":"John Smith","email":"test@example.com"}' | jq .
	@echo ""

## build: Build frontend for production
build:
	@echo "🏗️  Building frontend..."
	cd src && npm run build
	@echo "✅ Build complete!"

## deploy: Deploy to Vercel
deploy:
	@echo "🚀 Deploying to Vercel..."
	vercel --prod
	@echo "✅ Deployment complete!"

## import-csv: Import invite_list.csv to Supabase database
import-csv:
	@echo "📥 Importing CSV to database..."
	@echo ""
	@if [ ! -f "api/.env" ]; then \
		echo "⚠️  api/.env file not found!"; \
		echo "Please create it with your Supabase credentials:"; \
		echo "  SUPABASE_URL=your-project-url"; \
		echo "  SUPABASE_API_KEY=your-api-key"; \
		exit 1; \
	fi
	@set -a && . api/.env && set +a && cd api/tools && go run import-csv.go
	@echo ""
	@echo "✅ Import complete!"
