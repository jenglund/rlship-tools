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
- React Native Web (for web platform)
- Webpack (for web bundling)

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
│   └── mobile/              # Universal app (mobile + web)
│       ├── src/            # Application source code
│       ├── web/            # Web-specific assets
│       └── webpack.config.js # Web bundling configuration
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
# Install root dependencies
npm install

# Install mobile app dependencies
cd apps/mobile
npm install
```

2. Start the development server:
```bash
# For mobile development
npm start

# For web development
npm run web

# For iOS simulator
npm run ios

# For Android emulator
npm run android
```

The web version will be available at http://localhost:19006

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
# Run all tests
cd apps/mobile
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm test -- --coverage

# Run web-specific tests
npm test -- --testMatch="**/*.web.test.{js,jsx,ts,tsx}"
```

### Backend
```bash
cd backend
go test ./...
```

The backend includes a comprehensive testing infrastructure:

- **Database Testing**: Automatically creates and manages test databases
  - Each test gets a fresh database with the latest schema
  - Test databases are automatically cleaned up
  - Includes helper functions for creating test data

- **HTTP Testing**: Helpers for testing HTTP endpoints
  - Easy request execution and response validation
  - Mock authentication middleware
  - JSON request/response handling

- **Test Fixtures**: Helpers for generating test data
  - User creation
  - Relationship management
  - Activity list generation

Example test:
```go
func TestExample(t *testing.T) {
    // Set up test database
    db := testutil.SetupTestDB(t)
    defer testutil.TeardownTestDB(t, db)

    // Create test data
    user := testutil.CreateTestUser(t, db)
    relationship := testutil.CreateTestRelationship(t, db, []testutil.TestUser{user})
    
    // Test HTTP endpoints
    router := gin.New()
    req := testutil.TestRequest{
        Method: "GET",
        Path:   "/api/relationships",
        Header: map[string]string{
            "Authorization": "Bearer test-token",
        },
    }
    resp := testutil.ExecuteRequest(t, router, req)
    
    // Validate response
    expected := testutil.TestResponse{
        Code: http.StatusOK,
        Body: gin.H{"relationships": []string{relationship.ID.String()}},
    }
    testutil.CheckResponse(t, resp, expected)
}
```

## Building for Production

### Web Build
```bash
# Build the web version
npm run build:web

# The build output will be in apps/mobile/web-build/
```

### Mobile Build
Follow the Expo build instructions for iOS and Android:
```bash
# Build for iOS
eas build --platform ios

# Build for Android
eas build --platform android
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