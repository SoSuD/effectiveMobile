CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_people_gender
    ON people(gender);

CREATE INDEX IF NOT EXISTS idx_people_nationality
    ON people(nationality);

CREATE INDEX IF NOT EXISTS idx_people_age
    ON people(age);

CREATE INDEX IF NOT EXISTS idx_people_name_trgm
    ON people USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_people_surname_trgm
    ON people USING gin (surname gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_people_patronymic_trgm
    ON people USING gin (patronymic gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_people_nationality_gender
    ON people(nationality, gender);

CREATE INDEX IF NOT EXISTS idx_people_id_desc
    ON people(id DESC);