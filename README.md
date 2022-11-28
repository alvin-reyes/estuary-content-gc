# Estuary Content Table Clean up

- Queries old data and check pinning status
- Check shuttle hosts if content is available
- If content is not available, mark it as failed.

# Create the DB connection .env file

```
DB_NAME=
DB_HOST=
DB_USER=
DB_PASS=
DB_PORT=
```

# Install run
```
go mod tidy
go mod download
go run main.go
```
