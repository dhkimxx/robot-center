# Database Schema

The local PoC schema is created by the Go app at startup through GORM `AutoMigrate`.

Keep this directory for future production-grade SQL migrations only. Do not add local
bootstrap table DDL here unless it cannot be represented by the app startup migration
path.
