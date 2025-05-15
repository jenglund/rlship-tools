-- Drop indexes
DROP INDEX IF EXISTS idx_relationship_members_user_id;
DROP INDEX IF EXISTS idx_activity_lists_relationship_id;
DROP INDEX IF EXISTS idx_activities_list_id;
DROP INDEX IF EXISTS idx_activity_notes_activity_id;
DROP INDEX IF EXISTS idx_activity_notes_user_id;
DROP INDEX IF EXISTS idx_activity_events_activity_id;
DROP INDEX IF EXISTS idx_activity_events_relationship_id;
DROP INDEX IF EXISTS idx_activity_event_photos_event_id;
DROP INDEX IF EXISTS idx_interest_buttons_relationship_id;
DROP INDEX IF EXISTS idx_interest_button_activations_button_id;
DROP INDEX IF EXISTS idx_interest_button_activations_user_id;

-- Drop tables
DROP TABLE IF EXISTS interest_button_activations;
DROP TABLE IF EXISTS interest_buttons;
DROP TABLE IF EXISTS activity_event_photos;
DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS activity_notes;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS activity_lists;
DROP TABLE IF EXISTS relationship_members;
DROP TABLE IF EXISTS relationships;
DROP TABLE IF EXISTS users; 