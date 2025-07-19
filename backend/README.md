# Email Sorting App Backend

## Setup

1. Install PostgreSQL and create a database:
```sql
CREATE DATABASE email_sorting_app;
```

2. Copy environment variables:
```bash
cp .env.example .env
```

3. Update `.env` with your configuration:
   - Set your PostgreSQL connection string
   - Add your Google OAuth credentials (from Google Cloud Console)

4. Apply database schema:
```bash
psql -d email_sorting_app -f schema.sql
```

5. Run the application:
```bash
go run main.go
```

## Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Gmail API
4. Create OAuth 2.0 credentials
5. Add `http://localhost:8080/auth/callback` to authorized redirect URIs
6. Update `.env` with your client ID and secret

## API Endpoints

- `GET /auth/login` - Get Google OAuth URL
- `GET /auth/callback` - OAuth callback handler  
- `GET /accounts` - List connected accounts
- `GET /accounts/:id/emails` - Get emails for account
- `DELETE /accounts/:id` - Remove account
- `POST /auth/logout` - Sign out