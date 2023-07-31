set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS users (
       id INT PRIMARY KEY,
       nickname varchar(128) NOT NULL,
       email varchar(128) NOT NULL
    );
    CREATE TABLE IF NOT EXISTS ads (
       id SERIAL PRIMARY KEY,
       title varchar(512) NOT NULL,
       text varchar(4096) NOT NULL,
       author_id INT NOT NULL,
       published BOOLEAN NOT NULL,
       creation_date BIGINT,
       update_date BIGINT
    );
EOSQL