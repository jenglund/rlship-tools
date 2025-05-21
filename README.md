# RLship Tools aka "Tribe" (gonetribal.com)

This repository contains the Tribe application stack, including backend services and web interface. Tribe is a platform for managing and sharing activities with your tribes - whether they're couples, families, friend groups, or any other group of people who want to do things together.


See [KNOWNISSUES.md](KNOWNISSUES.md) for details on current issues and [FUTUREWORK.md](FUTUREWORK.md) for the complete roadmap.

## Project Structure

- **frontend**: React 19-based web application for Tribe
- **backend**: Backend services including API, models, and database connections
- **docs**: Documentation for Tribe
- **infrastructure**: Docker, Kubernetes, and other infrastructure configurations
- **shared**: Shared code and libraries used across the Tribe platform
- **tools**: Development and deployment tools

## Development Setup

### Backend Development

```bash
make dev-backend
```

### Frontend Development

```bash
make dev-frontend
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

or separately

```bash
make test-backend
make test-frontend
```

## Documentation

The following documentation files provide additional details about various aspects of the project. These files should be referenced, followed, and updated as strictly as the README itself.

- [**DESIGN.md**](DESIGN.md): Design principles, user types, terminology, and foundational concepts
- [**STATUS.md**](STATUS.md): Current project status and milestones
- [**KNOWNISSUES.md**](KNOWNISSUES.md): Known issues and their status
- [**FUTUREWORK.md**](FUTUREWORK.md): Roadmap for future development and implementation plans
- [**DEVGUIDANCE.md**](DEVGUIDANCE.md): Comprehensive development guidelines, best practices, and philosophies

## License

**MIT!** See the [**LICENSE**](LICENSE) file for more details.
