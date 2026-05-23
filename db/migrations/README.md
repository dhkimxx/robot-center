# Database Migrations

`deploy/docker-compose.yml` mounts this directory into the PostgreSQL container as `/docker-entrypoint-initdb.d`.

The SQL files are applied only when the PostgreSQL data volume is first created. During early P0 development, remove the `postgres-data` volume if the schema needs to be re-applied from scratch.
