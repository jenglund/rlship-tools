# Authentication Changes Task List

## Overview
The current authentication system is too restrictive, requiring specific dev user email patterns (`dev_user[0-9]+@gonetribal.com`). We need to make the following changes:

1. Allow all users to access the site in development mode (no more 401s for non-dev users)
2. Consolidate dev user checks to a single function in `dev_auth.go`
3. Define dev users as anyone with an email ending in `@gonetribal.com`
4. Keep admin panel access restricted to dev users only
5. Remove any pattern enforcement for dev users on the registration page

## Tasks

### Backend Changes
- [x] Modify `dev_auth.go` to remove the restrictive email pattern check
  - [x] Change `devEmailPattern` to check only for `@gonetribal.com` domain
  - [x] Update `VerifyIDToken` to accept all tokens in development mode
  - [x] Update `AuthMiddleware` to skip the dev user pattern check
  - [x] Create a new helper function `IsDevUser(email string) bool` that checks if the email ends with `@gonetribal.com`
- [x] Fix user creation and retrieval issues
  - [x] Update `RegisterUser` handler to properly handle the `firebase_uid` field
  - [x] Enhance `CreateOrGetDevUser` method to be more robust
  - [x] Ensure users are created properly for all authentication methods

### Frontend Changes
- [x] Update Login Screen to remove dev user pattern instruction
- [x] Update Registration Screen to remove any dev user pattern enforcement
- [x] Ensure `authService.js` properly sets auth headers for all users, not just dev users
- [x] Verify admin panel still correctly uses the `isAdmin` function to check for `@gonetribal.com` domain

### Testing
- [x] Fix 401 "User not found" error when trying to create a tribe
- [x] Fix 500 error when registering users via POST to `/api/users/register`
- [x] Test logging in with a non-dev user (e.g., `test@example.com`)
- [x] Test accessing protected routes with a non-dev user
- [x] Test accessing the admin panel with both dev and non-dev users
- [x] Test admin functions with a dev user

### Documentation
- [x] Update README.md to reflect the new authentication system 