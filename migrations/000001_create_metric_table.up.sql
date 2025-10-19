CREATE TABLE IF NOT EXISTS metric (
    "id" varchar(64) primary key, 
    "type" varchar(16) NOT NULL, 
    "delta" bigint, 
    "value" double precision, 
    "hash" varchar(255)
);
