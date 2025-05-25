DROP INDEX IF EXISTS idx_people_id_desc;
DROP INDEX IF EXISTS idx_people_nationality_gender;
DROP INDEX IF EXISTS idx_people_patronymic_trgm;
DROP INDEX IF EXISTS idx_people_surname_trgm;
DROP INDEX IF EXISTS idx_people_name_trgm;
DROP INDEX IF EXISTS idx_people_age;
DROP INDEX IF EXISTS idx_people_nationality;
DROP INDEX IF EXISTS idx_people_gender;

-- При желании можно оставить pg_trgm для других миграций, но если нужно, то:
DROP EXTENSION IF EXISTS pg_trgm;