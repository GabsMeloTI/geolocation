DROP INDEX IF EXISTS idx_states_search_vector;
DROP INDEX IF EXISTS idx_cities_search_vector;
DROP INDEX IF EXISTS idx_neighborhoods_search_vector;
DROP INDEX IF EXISTS idx_streets_search_vector;

ALTER TABLE states DROP COLUMN IF EXISTS search_vector;
ALTER TABLE cities DROP COLUMN IF EXISTS search_vector;
ALTER TABLE neighborhoods DROP COLUMN IF EXISTS search_vector;
ALTER TABLE streets DROP COLUMN IF EXISTS search_vector;

DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS streets;
DROP TABLE IF EXISTS neighborhoods;
DROP TABLE IF EXISTS cities;
DROP TABLE IF EXISTS states;
DROP TABLE IF EXISTS unique_ceps;
DROP EXTENSION IF EXISTS pg_trgm;
