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

## Database Migrations

### Development Guidelines

During development, we're keeping migrations simplified with a single migration file:

- **Use only the initial migration file** (`000001_initial_schema.up.sql`) for all schema changes
- Do NOT create additional migration files during development
- If you need to make schema changes, modify the initial migration file directly
- When you make changes to the schema, you can safely drop and recreate the database

```bash
# Drop and recreate the development database
make migrate-down
make migrate-up
```

This approach simplifies testing and development. Once we reach production, we'll implement proper incremental migrations.

## Known Issues

### Fixed Issues
- ✅ Activity Sharing Issues:
  - Fixed deduplication logic in GetSharedActivities
  - Fixed share expiration handling
  - Fixed unshare functionality
  - Added proper validation for private activity sharing - now correctly updates visibility

- ✅ **Tribe Repository Type Validation**: Fixed issues with tribe type validation in the repository layer. The code now properly initializes tribe objects with proper Type field and handles members correctly. This fixes the core tribe operations.
  - Modified the tribe repository methods to properly handle the Type field
  - Fixed the implementation of List, GetByID, GetUserTribes, GetByType, and Search methods
  - Restructured the code to avoid using the problematic getMembersWithTx function which was causing parse errors

- ✅ **List Repository Foreign Key Issues**: Fixed foreign key constraint issues in list repository tests:
  - Fixed TestListRepository_GetByID by creating a user before testing
  - Fixed TestListRepository_GetOwners by creating both a user and tribe before testing
  - Fixed TestListRepository_UnshareWithTribe by ensuring proper test setup with existing users and tribes
  - Fixed TestListRepository_GetListsBySource by properly setting the SyncStatus field

- ✅ **Database Schema Inconsistency Issue**:
  - Consolidated table names from "list_shares" to "list_sharing" to match migration schema
  - Added missing version field to list_sharing and activity_shares tables
  - Fixed references in test utilities to use consistent table names
  - Updated all repository code to use the correct table names
  - Ensured proper soft delete checks in queries (WHERE deleted_at IS NULL)
  - Modified UnshareWithTribe function to also remove the tribe as owner from the list_owners table

### Remaining Issues

#### Critical Issues:
None! All tests are now passing.

#### Areas for Improvement:
1. **Test Coverage**:
   - Current backend repository coverage is 48.8%
   - Should aim for at least 70-80% coverage

2. **Database Schema Improvements**:
   - Add unique constraint for tribe names
   - Consider adding additional indexes for performance
   - Add comprehensive validation at the database level

3. **API Improvements**:
   - Add comprehensive request validation
   - Implement rate limiting
   - Enhance error responses

## Current Focus
All backend tests are now passing! The main issues we resolved were:

1. Ensuring consistent table naming throughout the codebase (standardized on "list_sharing" instead of "list_shares")
2. Adding the missing version field to tables used by the trigger system
3. Fixing the UnshareWithTribe function to properly handle removing tribes as owners
4. Ensuring consistent use of soft delete checks in queries

Next steps should focus on increasing test coverage and implementing additional features.

## Future Work

- Increase test coverage for all components (current backend repository coverage: 48.8%)
- Implement database schema improvements (e.g., adding unique constraint for tribe names)
- Add comprehensive logging around database operations for better debugging
- Enhance API validation and error handling
- Implement rate limiting for API endpoints
- Add performance monitoring and optimization
- Develop additional features for the Tribe application

## License

See the LICENSE file for details. 