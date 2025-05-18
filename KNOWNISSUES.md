# Known Issues

This document tracks the current known issues in the Tribe project that need to be addressed.

## Frontend Issues

### `ConfigError` in Mobile App
- **Issue**: The mobile application is encountering a `ConfigError` when attempting to start. The exact cause is under investigation.
- **Suspected Packages**: Potentially outdated or misconfigured Expo-related packages like `@expo/config-types`, `@expo/metro-config`, or `jest-expo` despite `expo` and `react` being at target versions (53.0.0 and 19.1.0 respectively).
- **Status**: Under investigation
- **Priority**: High

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