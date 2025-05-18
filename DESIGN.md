# Design Principles and Concepts

This document outlines the design principles, user types, and foundational concepts that guide the development of the Tribe application.

## User Types and Terminology

### Dev Users vs Test Users

- **Dev Users**: Test accounts that developers can use to interact with the local application stack
  - Persist in the local database and should only be reset by running a developer script
  - Should have emails in the format `dev_user#@gonetribal.com` (e.g., `dev_user1@gonetribal.com`)
  - Currently, we have four dev users: `dev_user1`, `dev_user2`, `dev_user3`, and `dev_user4`

- **Test Users**: Test accounts used by automated testing (unit tests, end-to-end tests, etc.)
  - May be deleted and recreated as needed for testing and automation
  - Should have emails in the format `test_user#@gonetribal.com` or with descriptive names like `test_user_broken_sync_tester1@gonetribal.com`

## Fundamental Design Concepts

Tribe is designed to facilitate group activities and connections between people with various kinds of relationships.

- **Tribes**: Groups of people sharing activities and interests. While initially focused on romantic couples, the app is designed to support:
  - Couples planning date nights
  - Polyamorous relationships
  - Friend groups
  - Families
  - Roommates
  - Any other group of people who want to coordinate activities together

- **Lists**: Collections of places, activities, or things that tribes want to do together
  - Should support syncing with external sources (e.g., Google Maps lists)
  - Track which activities have been done and when
  - Help suggest new activities (e.g., places not visited in the last 3 months)

- **Interest Indicators**: Ways for tribe members to express availability or interest
  - Can connect with calendar entries (Google Calendar)
  - Allow for immediate/current interest indicators
  - Support scheduling future availability

## User Stories

### User Story #1: Basic Authentication and Tribe Viewing Flow

**As a user, I want to log in to the application and view my tribes and their members.**

1. **Flow**:
   - User starts the app or visits the site
   - User sees a lightweight splash screen with "Tribe" and a "Log In" button
   - User taps/clicks "Log In" and is taken to the login screen
   - During development, the login screen shows four "Login as dev_user#" buttons
   - After login, user lands on a screen listing all tribes they belong to
   - User can tap/click on a tribe to see all members of that tribe
   - User can navigate back to the tribes list

2. **Design Principles**:
   - Authentication will initially support dev users for development, but will eventually use Google authentication
   - UI should be lightweight and fast-loading
   - Navigation should be intuitive and follow platform conventions
   - User should always know which tribe they're viewing
   - Clear information hierarchy with tribes as the main organizing concept

## User Interface Design

1. **Display Names in Tribes**:
   - Users must have at least one display name (not NULL)
   - In the future, support per-tribe display names:
     - Users may want different display names in different tribes (e.g., real name in family tribe, gamertag in gaming tribe)
     - Some communities know people by different names (IG/TikTok usernames, nicknames, etc.)
     - Initially implement a single "default" display name
     - Plan to add tribe-specific display name overrides in future iterations

2. **Mobile-First Design**:
   - Optimize for mobile experiences first
   - Support responsive layouts for larger screens
   - Follow platform-specific UI/UX guidelines (iOS, Android, Web)
   - Use React Native Paper for consistent component styling

3. **Navigation Structure**:
   - Tab-based navigation for main app sections:
     - Home/Tribes
     - Lists
     - Interest Buttons
     - Profile
   - Stack navigation for drilling down into details
   - Clear back navigation and hierarchy awareness

4. **State Management**:
   - Centralized state for authentication
   - Tribe-specific state containers
   - Optimistic updates for responsive UI
   - Proper loading and error states 