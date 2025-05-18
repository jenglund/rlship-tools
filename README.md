# RLship Tools aka "Tribe" (gonetribal.com)

This repository contains the Tribe application stack, including backend services and web interface. Tribe is a platform for managing and sharing activities with your tribes - whether they're couples, families, friend groups, or any other group of people who want to do things together.

## Project Structure

- **frontend**: React 19-based web application for Tribe
- **backend**: Backend services including API, models, and database connections
- **docs**: Documentation for Tribe
- **infrastructure**: Docker, Kubernetes, and other infrastructure configurations
- **shared**: Shared code and libraries used across the Tribe platform
- **tools**: Development and deployment tools
- **ignore**: Deprecated code that is kept for reference purposes only

## Development Setup

### Backend Development

```bash
cd backend
make dev
```

### Frontend Development

```bash
cd frontend
npm install
npm start
```

### Development Authentication

During development, the backend supports a simplified authentication flow that doesn't require Firebase credentials. To use this:

1. Set the environment variable `ENVIRONMENT=development` or ensure the Firebase credentials file doesn't exist
2. The server will automatically enable development authentication mode
3. Use dev user accounts with email addresses matching the pattern `dev_user[0-9]+@gonetribal.com`

You can authenticate in one of two ways:
- Include a header `X-Dev-Email: dev_user1@gonetribal.com` in your requests
- Send a token in the format `dev:dev_user1@gonetribal.com` in the Authorization header

Example:
```bash
# Using X-Dev-Email header
curl -H "X-Dev-Email: dev_user1@gonetribal.com" http://localhost:8080/api/users/me

# Using Authorization header
curl -H "Authorization: Bearer dev:dev_user1@gonetribal.com" http://localhost:8080/api/users/me
```

## Testing

```bash
make test
```

## Documentation

The following documentation files provide additional details about various aspects of the project. These files should be referenced, followed, and updated as strictly as the README itself.

- [**DESIGN.md**](DESIGN.md): Design principles, user types, terminology, and foundational concepts
- [**STATUS.md**](STATUS.md): Current project status and milestones
- [**KNOWNISSUES.md**](KNOWNISSUES.md): Known issues and their status
- [**FUTUREWORK.md**](FUTUREWORK.md): Roadmap for future development and implementation plans
- [**DEVGUIDANCE.md**](DEVGUIDANCE.md): Comprehensive development guidelines, best practices, and philosophies

## Acknowledgements

This codebase would not have been possible without the significant contributions and assistance from:

- **Cursor**: The AI-powered IDE that dramatically improved coding efficiency
- **Claude**: For providing intelligent code analysis, suggestions, and debugging assistance
- **Gemini**: For additional AI-powered support in development

## Project Statistics

- **Total Files**: 140
- **Total Lines of Code**: 84,538
- **Total Size**: 2.79 MB

**Development Speed**: This codebase represents work that would typically take a team of 2-3 human developers approximately 3-6 months to build, test, and stabilize. The use of AI-assisted development has enabled us to reach this milestone in a fraction of the time while maintaining high code quality standards.

## License

See the LICENSE file for details.
