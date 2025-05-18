# Known Issues

This document tracks the current known issues in the Tribe project that need to be addressed.

## Frontend Issues

### Web Application Development Status
- **Update**: Web application has been configured to use React 19.1.0. 
- **Current Status**: Application development is proceeding with React 19, which provides modern features and performance improvements.
- **Priority**: Medium

### Webpack Dev Server Deprecation Warnings
- **Issue**: When running the web application, some deprecation warnings may appear related to webpack-dev-server:
  - `[DEP_WEBPACK_DEV_SERVER_CONSTRUCTOR]`: Using 'compiler' as first argument is deprecated
  - `[DEP_WEBPACK_DEV_SERVER_LISTEN]`: 'listen' method is deprecated
  - `[DEP_WEBPACK_DEV_SERVER_CLOSE]`: 'close' method is deprecated
- **Status**: These warnings don't affect functionality and will be automatically resolved when webpack updates in a future release.
- **Resolution Decision**: We've decided to accept these deprecation warnings rather than implement custom workarounds that may introduce other issues.
- **Priority**: Low

### TypeScript Integration
- **Issue**: The project currently uses JavaScript and needs TypeScript integration for better type safety.
- **Possible Solution**: Set up TypeScript and add type definitions for React 19 components.
- **Priority**: Medium

## Backend Issues

### Linting Issues
- **`govet shadow` errors**: The linter has identified 46 instances of variable shadowing across the backend codebase.
- **Resolution Decision**: We've decided to keep the `govet shadow` linter disabled for the time being. This is a low-impact issue that would require significant time and resources to address for minimal reward.

### Test Coverage Gaps
- Some repository methods have skipped tests due to transaction-related challenges:
  - `TestListRepository_GetUserLists/test_repository_method`
  - `TestListRepository_GetTribeLists/test_repository_method`
  - `TestListRepository_GetSharedLists`
  - `TestListRepository_ShareWithTribe/update_existing`
  - `TestListRepository_GetListShares/multiple_versions`
- **Status**: To be addressed in future test coverage improvements
- **Priority**: Medium 