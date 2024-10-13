# Web-Tester

This is a project to learn more about browser emulation.

Contained inside is a code that will:

1. Start a chrome browser headlessly
2. Navigate to a website
3. Capture every request and response
4. Input the results into a postgres database

If you wish to test, run the following:

```bash
docker compose up -d
```

```bash
go run cmd/main.go
```