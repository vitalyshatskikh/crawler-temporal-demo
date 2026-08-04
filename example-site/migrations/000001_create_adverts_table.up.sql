CREATE TABLE adverts (
    id          TEXT NOT NULL,
    region      TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    price       INTEGER NOT NULL,
    pub_date    TIMESTAMPTZ NOT NULL,
    version     INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (region, id)
);

CREATE INDEX idx_adverts_region_pub_date ON adverts (region, pub_date);
