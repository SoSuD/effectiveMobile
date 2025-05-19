CREATE TYPE gender_type AS ENUM ('male', 'female', 'unknown');

CREATE TABLE people (
                        id serial primary key,
                        name varchar(255),
                        surname varchar(255),
                        patronymic varchar(255),
                        age int,
                        gender gender_type,
                        nationality char(2)
)