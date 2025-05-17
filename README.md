# RLShip Tools

A collection of tools for managing relationships and tribes.

## Recent Changes

### Activity Sharing Improvements
- Fixed issues with activity sharing and ownership management
- Improved handling of expired shares
- Added proper owner tracking for shared activities
- Fixed deduplication of shared activities in query results

## Known Issues

### Activity Sharing
- Need to add tests for the new ownership functionality in activity sharing
- Consider adding batch operations for sharing activities with multiple tribes
- Consider adding activity share history tracking

### List Repository
- Consider aligning list sharing implementation with activity sharing improvements
- Add support for temporary/guest access to lists

## Future Work

### Activity Management
- Add support for recurring activities
- Implement activity categories and tags
- Add support for activity recommendations based on history

### List Management
- Add support for list templates
- Implement list version history
- Add support for collaborative list editing

### Tribe Management
- Add support for tribe hierarchies
- Implement tribe access levels and roles
- Add support for tribe activity feeds

## Development

### Prerequisites
- Go 1.19 or later
- PostgreSQL 13 or later
- Docker (optional)

### Setup
1. Clone the repository
2. Copy `.env.example` to `.env` and update the values
3. Run `make setup` to install dependencies
4. Run `make migrate` to set up the database
5. Run `make test` to verify the setup

### Testing
- Run `make test` to run all tests
- Run `make test-coverage` to run tests with coverage report
- Run `make test-integration` to run integration tests

### Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

See [LICENSE](LICENSE) for details. 