# RLShip Tools

A comprehensive relationship management application that helps couples and polyamorous relationships connect, plan activities, and share experiences together. The app supports multiple platforms including mobile web, desktop web, Android, and iOS.

## Features

- **Multi-Platform Support**
  - Mobile Web (Primary Focus)
  - Desktop Web
  - Android App
  - iOS App

- **Relationship Management**
  - Support for couples and polyamorous relationships
  - Relationship timeline and history
  - Photo sharing and memories

- **Activity Management**
  - Multiple customizable activity lists (restaurants, hikes, entertainment, etc.)
  - Public and private notes for each activity
  - Ratings and reviews
  - Visit history tracking
  - Photo attachments

- **Interest Button System**
  - Create custom interest buttons for any activity
  - Configure duration (minutes to weeks)
  - Loud (with notifications) or quiet mode
  - Optional early cancellation
  - Perfect for spontaneous activities or planning ahead

- **Smart Features**
  - Activity suggestions based on preferences and history
  - Historical activity queries
  - Time-based filtering (e.g., "not visited in 6 months")
  - Smart scheduling suggestions

## Tech Stack

### Frontend
- Expo / React Native
- TypeScript
- React Navigation
- React Native Paper
- Expo Push Notifications
- Firebase Authentication

### Backend
- Go
- PostgreSQL
- Redis
- Google Cloud Platform
  - Cloud Run / GKE
  - Cloud Storage
  - Cloud SQL

## Project Structure

```
rlship-tools/
├── apps/
│   ├── mobile/              # Expo/React Native application
│   └── web/                 # Web-specific optimizations
├── backend/                 # Go API server
│   ├── cmd/                 # Entry points
│   ├── internal/           # Internal packages
│   ├── pkg/                # Public packages
│   └── api/                # API definitions
├── infrastructure/         # Infrastructure as Code (Terraform)
├── shared/                # Shared types and utilities
├── tools/                 # Development tools and scripts
└── docs/                  # Documentation
```

## Development Setup

### Prerequisites

- Node.js (v18+)
- Go 1.21+
- Expo CLI
- Docker
- Google Cloud SDK
- PostgreSQL 17
  - On macOS with Homebrew: `brew install postgresql@17`
  - After installation:
    ```bash
    # Add PostgreSQL binaries to your PATH (add to ~/.zshrc or ~/.bashrc for permanence)
    export PATH="/usr/local/opt/postgresql@17/bin:$PATH"
    
    # Start PostgreSQL service
    brew services start postgresql@17
    ```
- Redis

### Frontend Setup

1. Install dependencies:
```bash
cd apps/mobile
npm install
```

2. Start the development server:
```bash
npm start
```

### Backend Setup

1. Install Go dependencies:
```bash
cd backend
go mod download
```

2. Set up the database:
```bash
# Ensure PostgreSQL service is running
brew services list | grep postgresql

# Create a PostgreSQL database (if PATH is set correctly)
createdb rlship

# Alternative command if createdb is not in PATH
/usr/local/opt/postgresql@17/bin/createdb rlship

# Run migrations
go run cmd/migrate/main.go up
```

3. Start the server:
```bash
go run cmd/api/main.go
```

### Environment Variables

Create a `.env` file in both the frontend and backend directories. Example variables needed:

```env
# Frontend (.env in apps/mobile)
API_URL=http://localhost:8080
FIREBASE_API_KEY=your_api_key
FIREBASE_PROJECT_ID=your_project_id

# Backend (.env in backend)
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=rlship
DB_USER=postgres
DB_PASSWORD=postgres
REDIS_URL=localhost:6379
GCP_PROJECT_ID=your_project_id
```

## Testing

### Frontend
```bash
cd apps/mobile
npm test
```

### Backend
```bash
cd backend
go test ./...
```

## Deployment

The application is designed to be deployed on Google Cloud Platform:

1. Frontend is deployed as a static website on Cloud Storage
2. Backend API runs on Cloud Run or GKE
3. Database runs on Cloud SQL
4. Redis runs on Memorystore
5. Media files are stored in Cloud Storage

Detailed deployment instructions can be found in the [deployment guide](docs/deployment.md).

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 