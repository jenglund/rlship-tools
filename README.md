# RLship Tools aka "Tribe" (gonetribal.com)

This repository contains the Tribe application stack, including backend services, mobile app, and web interface. Tribe is a platform for managing and sharing activities with your tribes - whether they're couples, families, friend groups, or any other group of people who want to do things together.

## Project Structure

- **apps**: Client applications (mobile and web interfaces for Tribe)
- **backend**: Backend services including API, models, and database connections
- **docs**: Documentation for Tribe
- **infrastructure**: Docker, Kubernetes, and other infrastructure configurations
- **shared**: Shared code and libraries used across the Tribe platform
- **tools**: Development and deployment tools

## Development Setup

### Backend Development

```bash
cd backend
make dev
```

### Mobile Development

```bash
cd apps/mobile
npm install
npm start
```

## Testing

```bash
make test
```

## Known Issues

### Fixed Issues
- ✅ Activity Sharing Issues:
  - Fixed deduplication logic in GetSharedActivities
  - Fixed share expiration handling
  - Fixed unshare functionality
  - Added proper validation for private activity sharing - now correctly updates visibility

### Remaining Issues

#### List Repository Issues:
- TestListRepository_GetByID/success failing with foreign key constraint error
- TestListRepository_GetOwners failing with null value constraint for owner_id
- TestListRepository_GetListsBySource failing with validation error for sync status
- TestListRepository_UnshareWithTribe failing with null value constraint for owner_id

#### Tribe Repository Issues:
- Various failures in tribe repository tests, including:
  - Error with duplicate name check
  - Errors related to tribe creation with invalid type
  - Issues with tribe member operations

## Future Work

- Fix the remaining list repository and tribe repository issues
- Improve error handling for external sources in Sync Service
- Fix concurrent operation handling in repositories
- Add proper state transition validation
- Enhance test coverage for all components

## License

See the LICENSE file for details. 