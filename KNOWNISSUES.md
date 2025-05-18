# Known Issues

This document tracks the current known issues in the Tribe project that need to be addressed.

## Frontend Issues

### Mobile App Development Status
- **Update**: Mobile application has been configured to use Expo 50.0.6 with React 18.2.0. 
- **Prior Issue**: Initially attempted to configure with Expo 53 but encountered compatibility issues.
- **Current Status**: Application development is proceeding with Expo 50, which has better compatibility with the project requirements.
- **Priority**: Medium

### Webpack Dev Server Deprecation Warnings
- **Issue**: When running the web application with Expo, several deprecation warnings appear related to webpack-dev-server:
  - `[DEP_WEBPACK_DEV_SERVER_CONSTRUCTOR]`: Using 'compiler' as first argument is deprecated
  - `[DEP_WEBPACK_DEV_SERVER_LISTEN]`: 'listen' method is deprecated
  - `[DEP_WEBPACK_DEV_SERVER_CLOSE]`: 'close' method is deprecated
- **Status**: These warnings don't affect functionality and will be automatically resolved when Expo updates its webpack configuration in a future release.
- **Resolution Decision**: We've decided to accept these deprecation warnings rather than implement custom workarounds that may introduce other issues.
- **Priority**: Low

### TypeScript Errors in Navigation Components
- **Issue**: The App.tsx file has TypeScript errors related to React Navigation components: `Navigator` components from react-navigation can't be used as JSX elements.
- **Affected Components**: `NavigationContainer`, `Stack.Navigator`, `Tab.Navigator`, and `TribesStack.Navigator`.
- **Possible Solution**: Update TypeScript and React Navigation type definitions to be compatible with our current setup.
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