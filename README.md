# Apso Go Service Template

The GitHub template behind every Apso Go backend: Gin, GORM, goose migrations, and PostgreSQL, with Swagger documentation via swaggo. `apso init --name my-api --language go` clones this repository, and `apso generate` turns your `.apsorc` schema into models, handlers, routes, and services inside it. The generated code is yours: Apache-2.0 licensed, no Apso SDK, no runtime dependency.

New to Apso? The [quickstart](https://docs.apso.ai/get-started/quickstart) goes from schema to running API in about ten minutes. All three language templates are listed at [apso.ai/templates](https://apso.ai/templates).

## Project Structure

```
.
├── cmd/
│   └── main.go              # Application entry point
├── autogen/                  # Auto-generated code (DO NOT EDIT)
│   ├── dto/                  # Data transfer objects
│   ├── handlers/             # HTTP request handlers
│   ├── models/               # Database models
│   │   └── enums/            # Enum types
│   ├── routes/               # Route registration
│   └── services/             # Business logic
├── extensions/               # Custom/manual code
│   └── middleware/           # Custom middleware
├── docs/                     # Generated Swagger docs
├── .apsorc                   # APSO CLI configuration
├── .env.example              # Environment variables template
├── Dockerfile                # Docker build configuration
├── go.mod                    # Go module definition
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL
- Docker (optional)

### Local Development

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Update `.env` with your database credentials

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Generate Swagger docs:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   swag init -g cmd/main.go -o docs
   ```

5. Run the server:
   ```bash
   go run cmd/main.go
   ```

### Docker

Build and run with Docker:

```bash
docker build -t my-api .
docker run -p 3000:3000 --env-file .env my-api
```

## API Documentation

Swagger documentation is available at: `http://localhost:3000/_docs`

## APSO CLI Integration

This template is designed to work with the APSO CLI. To add entities:

```bash
apso scaffold
```

The CLI will generate:
- Models in `autogen/models/`
- DTOs in `autogen/dto/`
- Handlers in `autogen/handlers/`
- Services in `autogen/services/`
- Routes in `autogen/routes/`

## Extending the Template

Add custom code in the `extensions/` folder:
- Custom middleware in `extensions/middleware/`
- Create additional folders as needed for your custom logic

**Important:** Never modify files in the `autogen/` folder as they will be overwritten by the APSO CLI.

## License

Apache License 2.0. See the [LICENSE](LICENSE) file.
