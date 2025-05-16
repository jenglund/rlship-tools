# Tribe - Social Activity Planning App

Tribe is a social activity planning application designed to help groups of people (tribes) share and plan activities together. Whether it's couples planning date nights, roommates coordinating house activities, or friend groups organizing outings, Tribe makes it easy to collaborate on and track shared activities.

## Core Features

- **Tribes**: Groups of people who share activities and plans together
- **Lists**: Shared collections of activities, places, or plans
- **Location Integration**: Sync with Google Maps lists for place suggestions
- **Availability Management**: Track when tribe members are free/interested in activities
- **Smart Suggestions**: Get activity recommendations based on history and preferences

## Project Structure

- `apps/`
  - `mobile/`: React Native mobile application
  - `web/`: Web application frontend
- `backend/`: Go backend service
  - `api/`: API definitions and handlers
  - `internal/`: Core business logic and models
  - `migrations/`: Database schema migrations
- `infrastructure/`: Deployment and infrastructure configuration
- `shared/`: Shared types and utilities
- `docs/`: Project documentation

## Development Setup

1. Clone the repository
2. Install dependencies:
   ```bash
   make setup
   ```
3. Set up environment variables (see .env.example)
4. Start the development environment:
   ```bash
   make dev
   ```

## Testing

Run tests with:
```bash
make test
```

Ensure test coverage remains above 90% for all new code.

## Known Issues

1. Need to implement proper data models for:
   - Tribes and tribe membership
   - Activity lists and items
   - Location integration with Google Maps
   - Availability tracking
   - User preferences

2. Database schema needs updates for:
   - Proper foreign key constraints
   - Indexes for performance
   - Soft delete support
   - Audit trails

3. Missing validation for:
   - Tribe membership rules
   - List sharing permissions
   - Activity data
   - Location data

## Future Work

1. Core Features Implementation:
   - Create and manage tribes
   - Share and sync activity lists
   - Google Maps integration
   - Availability tracking system
   - Smart activity suggestions

2. Infrastructure:
   - Set up CI/CD pipelines
   - Configure monitoring and logging
   - Implement caching layer
   - Add rate limiting

3. Testing:
   - Add integration tests
   - Set up end-to-end testing
   - Implement performance testing
   - Add security testing

4. Documentation:
   - API documentation
   - Development guides
   - Deployment procedures
   - User guides

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details. 