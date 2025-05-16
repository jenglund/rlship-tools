# Type System Documentation

## Overview

This document outlines the type system used throughout the Tribe application, focusing on consistency between the model layer and database layer.

## Core Types

### String-Based Enums

1. VisibilityType
   - Values: private, shared, public
   - Usage: Controls resource visibility
   - DB Type: VARCHAR(10)
   - Validation: Required, must be valid enum value

2. OwnerType
   - Values: user, tribe
   - Usage: Defines resource ownership type
   - DB Type: VARCHAR(10)
   - Validation: Required, must be valid enum value

3. ListType
   - Values: general, location, activity, interest, google_map
   - Usage: Defines list type
   - DB Type: VARCHAR(20)
   - Validation: Required, must be valid enum value

4. ListSyncStatus
   - Values: none, synced, pending, conflict
   - Usage: Tracks list synchronization state
   - DB Type: VARCHAR(10)
   - Validation: Required, must be valid enum value

5. TribeType
   - Values: couple, polycule, friends, family, roommates, coworkers, custom
   - Usage: Defines tribe type
   - DB Type: VARCHAR(20)
   - Validation: Required, must be valid enum value

6. MembershipType
   - Values: full, limited, guest
   - Usage: Defines tribe membership level
   - DB Type: VARCHAR(10)
   - Validation: Required, must be valid enum value

7. AuthProvider
   - Values: google, email, phone
   - Usage: Authentication provider type
   - DB Type: VARCHAR(10)
   - Validation: Required, must be valid enum value

### Complex Types

1. JSONMap
   - Go Type: map[string]interface{}
   - DB Type: JSONB
   - Validation: Must be valid JSON
   - Implementation: Needs proper Scan/Value methods

2. Metadata
   - Go Type: map[string]interface{}
   - DB Type: JSONB
   - Validation: Must be valid JSON
   - Implementation: Needs proper type conversion handling

### Nullable Fields

The following rules apply to nullable fields:

1. Use pointer types (*string, *int, etc.) for:
   - Optional fields that can be NULL in database
   - Fields that need three-state logic (unset/set-to-null/set-to-value)

2. Use non-pointer types for:
   - Required fields
   - Fields with default values
   - Boolean flags (use false as default)

## Known Issues and Fixes

1. JSONMap Implementation
   - Current: Stub implementation of Scan/Value
   - Fix: Implement proper JSON handling with type conversion

2. Metadata Type Consistency
   - Current: Inconsistent handling between models
   - Fix: Standardize on JSONMap type with proper validation

3. Enum Type Safety
   - Current: String-based with runtime validation
   - Fix: Add compile-time type checking where possible

4. Nullable Field Consistency
   - Current: Mixed usage of pointers
   - Fix: Apply consistent rules based on field semantics

## Database Constraints

1. Enum Types
   ```sql
   CREATE TYPE visibility_type AS ENUM ('private', 'shared', 'public');
   CREATE TYPE owner_type AS ENUM ('user', 'tribe');
   -- etc.
   ```

2. Check Constraints
   ```sql
   ALTER TABLE lists ADD CONSTRAINT valid_sync_status 
       CHECK (sync_status IN ('none', 'synced', 'pending', 'conflict'));
   ```

## Implementation Plan

1. Database Layer
   - [ ] Create proper enum types in database
   - [ ] Add check constraints for enum values
   - [ ] Update existing tables to use enum types

2. Model Layer
   - [ ] Implement proper JSONMap scanning/value methods
   - [ ] Standardize nullable field usage
   - [ ] Add compile-time type checking for enums
   - [ ] Update validation methods

3. Repository Layer
   - [ ] Update type conversions for enum types
   - [ ] Implement proper JSON handling
   - [ ] Add type assertion checks

4. Testing
   - [ ] Add type conversion tests
   - [ ] Add constraint violation tests
   - [ ] Add JSON handling tests 