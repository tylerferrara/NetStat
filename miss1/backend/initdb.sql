-- DEFINE TABLES

CREATE TABLE election (
    id SERIAL PRIMARY KEY,
    start_time TIMESTAMP,
    end_time TIMESTAMP
);

CREATE TABLE candidate (
    id SERIAL PRIMARY KEY,
    fk_election integer REFERENCES election(id),
    name VARCHAR(246)
);

CREATE TABLE citizen (
    id SERIAL PRIMARY KEY,
    SSN VARCHAR(6),
    DOB VARCHAR(246),
    is_registered BOOLEAN,
    has_voted BOOLEAN,
    UNIQUE(SSN)
);

CREATE TABLE vote (
    id SERIAL PRIMARY KEY,
    fk_election integer REFERENCES election(id),
    fk_citizen integer REFERENCES citizen(id),
    fk_candidate integer REFERENCES candidate(id),
    vote_time TIMESTAMP
);

-- GENERATE AN ELECTION

INSERT INTO election (start_time, end_time) VALUES (timestamp '2016-01-13 17:38:42', timestamp '2023-11-01 17:38:42');

-- GENERATE CANDIDATES

INSERT INTO candidate (fk_election, name) VALUES (1, 'Minushka');
INSERT INTO candidate (fk_election, name) VALUES (1, 'Zach');

-- GENERATE CITIZENS

INSERT INTO citizen (SSN, DOB, is_registered, has_voted) VALUES ('111110', '12-10-1991', TRUE, FALSE);
INSERT INTO citizen (SSN, DOB, is_registered, has_voted) VALUES ('111111', '01-21-1977', FALSE, FALSE);
INSERT INTO citizen (SSN, DOB, is_registered, has_voted) VALUES ('111112', '06-07-1986', FALSE, FALSE);
