-- Drop indexes
DROP INDEX IF EXISTS idx_interest_button_activations_user_id;
DROP INDEX IF EXISTS idx_interest_button_activations_button_id;
DROP INDEX IF EXISTS idx_interest_buttons_relationship_id;
DROP INDEX IF EXISTS idx_relationship_members_user_id;

-- Drop tables
DROP TABLE IF EXISTS interest_button_activations;
DROP TABLE IF EXISTS interest_buttons;
DROP TABLE IF EXISTS relationship_members;
DROP TABLE IF EXISTS relationships;
DROP TABLE IF EXISTS users; 