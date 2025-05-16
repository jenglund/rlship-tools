# List Synchronization

This document describes the synchronization system for lists in the Tribe application, including workflows and conflict resolution strategies.

## Overview

The sync system allows lists to be synchronized with external sources (e.g., Google Maps lists). It manages the synchronization state and handles conflicts that may arise during sync operations.

## Sync States

Lists can be in one of the following sync states:

- `none`: No sync configured
- `pending`: Changes pending sync
- `synced`: In sync with external source
- `conflict`: Conflicts detected

## State Transitions

Valid state transitions are:

1. `none` → `pending`: When sync is first configured
2. `pending` → `synced`: When sync completes successfully
3. `pending` → `conflict`: When conflicts are detected during sync
4. `synced` → `pending`: When local changes are made
5. `synced` → `conflict`: When remote changes conflict with local state
6. `conflict` → `pending`: When conflicts are resolved manually
7. `conflict` → `synced`: When conflicts are auto-resolved
8. Any state → `none`: When sync is disabled

## Sync Workflow

### Configuring Sync

1. User selects a list to sync with an external source
2. System validates the source and sync ID
3. List transitions to `pending` state
4. Initial sync is triggered

### Normal Operation

1. System periodically checks for changes in external source
2. When changes are detected:
   - If no conflicts, changes are applied and list remains `synced`
   - If conflicts detected, list transitions to `conflict` state
3. When local changes are made:
   - List transitions to `pending`
   - Changes are queued for next sync

### Conflict Resolution

When conflicts are detected, the system:

1. Creates a conflict record with:
   - Local data
   - Remote data
   - Conflict type
   - Affected item (if applicable)
2. Transitions list to `conflict` state
3. Notifies relevant users

#### Resolution Strategies

1. **Manual Resolution**
   - User reviews conflicting changes
   - Chooses to use either local or remote version
   - Can modify data to create a merged version
   - System applies resolution and transitions to `pending`

2. **Auto-Resolution Rules**
   - Last-write-wins: Use most recent change
   - Source-priority: External source always wins
   - Local-priority: Local changes always win
   - Merge-fields: Combine non-conflicting field changes

#### Conflict Types

1. **Item Updates**
   - Name changes
   - Description changes
   - Metadata changes

2. **List Structure**
   - Item additions
   - Item removals
   - Order changes

3. **List Settings**
   - Visibility changes
   - Configuration changes

## Error Handling

The system handles various error conditions:

1. **Configuration Errors**
   - Invalid sync source
   - Missing sync ID
   - Invalid sync status

2. **State Transition Errors**
   - Invalid state transitions
   - Sync already enabled/disabled

3. **External Source Errors**
   - Source unavailable
   - Timeout
   - API errors

## Best Practices

1. **Minimizing Conflicts**
   - Sync frequently to reduce divergence
   - Lock items during edits
   - Use optimistic locking for concurrent modifications

2. **Handling Failures**
   - Implement retry logic with backoff
   - Log all sync operations
   - Monitor sync health

3. **User Experience**
   - Show sync status clearly
   - Provide clear conflict resolution UI
   - Allow bulk conflict resolution
   - Maintain audit trail of resolutions 