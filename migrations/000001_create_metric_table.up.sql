CREATE TABLE IF NOT EXISTS metric (
    "id" varchar(64) primary key, 
    "type" varchar(16) NOT NULL, 
    "delta" integer, 
    "value" real, 
    "hash" varchar(255)
);
