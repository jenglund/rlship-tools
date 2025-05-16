# RLShip Tools

A comprehensive relationship management application that helps tribes (couples and polyamorous relationships) connect, plan activities, and share experiences together. The app supports multiple platforms including mobile web, desktop web, Android, and iOS.

## Features

- **Multi-Platform Support**
  - Mobile Web (Primary Focus)
  - Desktop Web
  - Android App
  - iOS App

- **Tribe Management**
  - Support for couples and polyamorous relationships
  - Flexible membership types (full members and guests)
  - Time-limited guest access
  - Tribe timeline and history
  - Photo sharing and memories

- **Activity Management**
  - Multiple activity types (locations, interests, lists, general activities)
  - Public and private visibility settings
  - Activity sharing between tribes
  - Customizable metadata for each activity type
  - Visit history tracking
  - Photo attachments
  - Multiple ownership (users and tribes can own activities)

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
│   │   ├── api/            # API handlers and middleware
│   │   ├── models/         # Domain models
│   │   ├── repository/     # Data access layer
│   │   └── testutil/       # Testing utilities
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
- Expo CLI (`npm install -g expo-cli`)
- Docker Desktop
- Google Cloud SDK
- PostgreSQL 17
  - On macOS with Homebrew: `brew install postgresql@17`
  - After installation:
    ```bash
    # Add PostgreSQL binaries to your PATH (add to ~/.zshrc or ~/.bashrc for permanence)
    export PATH="/usr/local/opt/postgresql@17/bin:$PATH"
    
    # Start PostgreSQL service
    brew services start postgresql@17
    
    # Verify PostgreSQL is running
    brew services list | grep postgresql
    ```
- Redis
  - On macOS with Homebrew: `brew install redis`
  - After installation:
    ```bash
    # Start Redis service
    brew services start redis
    
    # Verify Redis is running
    brew services list | grep redis
    ```

### Backend Setup

1. Clone the repository and navigate to the backend directory:
```bash
git clone https://github.com/yourusername/rlship-tools.git
cd rlship-tools/backend
```

2. Install Go dependencies:
```bash
go mod download
go mod verify
```

3. Set up environment variables:
```bash
# Create .env file in backend directory
cp .env.example .env

# Edit the .env file with your configuration
# Required variables:
# - PORT=8080
# - DB_HOST=localhost
# - DB_PORT=5432
# - DB_NAME=rlship
# - DB_USER=postgres
# - DB_PASSWORD=postgres
# - REDIS_URL=localhost:6379
# - GCP_PROJECT_ID=your_project_id
```

4. Set up the database:
```bash
# Create a PostgreSQL database
createdb rlship

# Run migrations
go run cmd/migrate/main.go up

# Verify migrations
go run cmd/migrate/main.go version
```

5. Start the server:
```bash
go run cmd/api/main.go
```

The API server will be available at http://localhost:8080

### Kubernetes Local Development

For local development with Kubernetes:

1. Enable Kubernetes in Docker Desktop:
   - Open Docker Desktop
   - Go to Settings/Preferences
   - Select "Kubernetes"
   - Check "Enable Kubernetes"
   - Click "Apply & Restart"

2. Install required tools:
```bash
# Install kubectl
brew install kubectl

# Install Helm
brew install helm

# Install ingress-nginx controller
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx
```

3. Deploy the application:
```bash
# Deploy all services
./infrastructure/scripts/deploy-local.sh
```

4. Access the services:
   - Web Frontend: http://localhost
   - API: http://localhost/api
   - Lists Service: http://localhost/lists
   - Activities Service: http://localhost/activities
   - Interests Service: http://localhost/interests

5. View service status:
```bash
# Check pod status
kubectl get pods -n tribe

# Check service status
kubectl get services -n tribe

# Check ingress status
kubectl get ingress -n tribe

# View logs for a specific pod
kubectl logs -n tribe <pod-name>
```

For more details about the Kubernetes setup, see the [infrastructure/README.md](infrastructure/README.md) file.

### Frontend Setup

1. Navigate to the mobile app directory:
```bash
cd apps/mobile
```

2. Install dependencies:
```bash
npm install
```

3. Set up environment variables:
```bash
# Create .env file
cp .env.example .env

# Edit the .env file with your configuration
# Required variables:
# - API_URL=http://localhost:8080
# - FIREBASE_API_KEY=your_api_key
# - FIREBASE_PROJECT_ID=your_project_id
```

4. Start the development server:
```bash
# For web development
npm run web

# For iOS simulator (requires Xcode)
npm run ios

# For Android emulator (requires Android Studio)
npm run android

# For Expo development client
npm start
```

The web version will be available at http://localhost:19006

## Testing

### Backend Tests

The backend includes comprehensive test coverage with the following features:
- Automated test database management
- HTTP endpoint testing
- Mock authentication
- Test fixtures and data generation

Run the tests:
```bash
cd backend

# Run all tests
go test ./...

# Run tests with coverage and generate HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test package
go test ./internal/api/handlers/...

# Run tests with verbose output
go test -v ./...

# Run tests with race condition detection
go test -race ./...
```

View test coverage in terminal:
```bash
go test -cover ./...
```

### Frontend Tests

Run the frontend tests:
```bash
cd apps/mobile

# Run all tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm test -- --coverage

# Run specific tests
npm test -- -t "test name"

# Update snapshots
npm test -- -u
```

## Local Development Stack

To run the complete stack locally:

1. Start the database and cache:
```bash
# Start PostgreSQL
brew services start postgresql@17

# Start Redis
brew services start redis
```

2. Start the backend server:
```bash
cd backend
go run cmd/api/main.go
```

3. Start the frontend development server:
```bash
cd apps/mobile
npm run web  # For web development
# OR
npm start    # For mobile development
```

4. Access the applications:
- Web: http://localhost:19006
- API: http://localhost:8080
- Mobile: Use Expo Go app or simulators

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass and coverage is maintained
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 