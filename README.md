docker-compose up -d

# Important:

I have changed port numbers in docker-compose.yml file because those ports are in use for another software on my laptop.

# How to run reporting API?

```bash
PORT=9090 go run cmd/reporting_api/main.go
```

# How to run message processor?

```bash
go run cmd/message_processor/main.go 
```

# How to run API?

```bash
PORT=9080 go run cmd/api/main.go
```