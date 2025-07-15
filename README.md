# Flight Booking Routes API

http://flight-booking-lb-1909835177.eu-central-1.elb.amazonaws.com/api/v1/routes
A cloud-native flight booking aggregator system that provides a unified API for flight route information from multiple providers. Built with Go, following clean architecture principles, and ready for deployment in any cloud environment.

## Features

- **Multi-Provider Aggregation**: Fetches flight routes from multiple providers concurrently
- **Intelligent Caching**: Configurable TTL-based caching to reduce provider API calls
- **Health Monitoring**: Comprehensive health checks for all providers
- **Clean Architecture**: Separated concerns with handlers, use cases, and providers
- **Dependency Injection**: Uses Uber FX for robust dependency management
- **OpenAPI Documentation**: Auto-generated swagger documentation
- **Production Ready**: Includes logging, metrics, graceful shutdown, and error handling
- **Cloud Native**: Containerized with Docker

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│   API Gateway   │───▶│    Handlers     │
│  (Swagger UI)   │    │  (Rate Limit)   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                         │
                                                         ▼
                                               ┌─────────────────┐
                                               │   Use Cases     │
                                               │ (Business Logic)│
                                               └─────────────────┘
                                                         │
                                                         ▼
                                               ┌─────────────────┐
                                               │ Provider Manager│
                                               │   (Aggregator)  │
                                               └─────────────────┘
                                                         │
                                      ┌─────────────────┼─────────────────┐
                                      ▼                 ▼                 ▼
                                ┌──────────┐      ┌──────────┐      ┌──────────┐
                                │Provider1 │      │Provider2 │      │Cache     │
                                │          │      │          │      │Service   │
                                └──────────┘      └──────────┘      └──────────┘
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker (optional)
- Task (optional, for build automation)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/your-org/flight-booking.git
cd flight-booking
```

2. Install dependencies:
```bash
go mod download
```

3. Install development tools:
```bash
task install-tools
```

### Running the Application

#### Option 1: Using Task (Recommended)
```bash
# Generate code and run
task generate
task run

# Or for development with hot reload
task dev
```

#### Option 2: Using Go directly
```bash
# Generate models from OpenAPI spec
go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -generate types -package models api/openapi.yaml > internal/generated/models/types.go

# Run the application
go run cmd/server/main.go
```

#### Option 3: Using Docker
```bash
# Build and run with Docker
docker build -t flight-booking .
docker run -p 8080:8080 flight-booking

# Or use Docker Compose
docker-compose up
```

### API Documentation

Once the application is running, access the API documentation at:
- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI Spec: `http://localhost:8080/swagger/doc.json`

## API Endpoints

### Get Flight Routes
```http
GET /api/v1/routes?airline=AA&sourceAirport=JFK&destinationAirport=LAX&maxStops=2
```

**Response:**
```json
{
  "data": [
    {
      "airline": "AA",
      "sourceAirport": "JFK",
      "destinationAirport": "LAX",
      "codeShare": "Y",
      "stops": 0,
      "equipment": "737",
      "provider": "provider1"
    }
  ],
  "metadata": {
    "totalCount": 1,
    "providersUsed": ["provider1", "provider2"],
    "cacheHit": false,
    "timestamp": "2023-12-01T10:30:00Z"
  }
}
```

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-12-01T10:30:00Z",
  "providers": {
    "provider1": "healthy",
    "provider2": "healthy"
  }
}
```

## Configuration

The application supports multiple configuration methods:

### 1. Configuration File (config.yaml)
```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"

log:
  level: "info"
  format: "json"

cache:
  enabled: true
  default_ttl: "5m"
  cleanup_interval: "10m"

providers:
  provider1:
    enabled: true
    base_url: "https://api.provider1.com"
    timeout: "10s"
    retries: 3
  provider2:
    enabled: true
    base_url: "https://api.provider2.com"
    timeout: "10s"
    retries: 3
```

### 2. Environment Variables
All configuration can be overridden with environment variables using the `FLIGHT_BOOKING_` prefix:

```bash
export FLIGHT_BOOKING_SERVER_PORT=8080
export FLIGHT_BOOKING_LOG_LEVEL=debug
export FLIGHT_BOOKING_CACHE_ENABLED=true
export FLIGHT_BOOKING_PROVIDERS_PROVIDER1_BASE_URL=https://api.provider1.com
```

## Development

### Project Structure
```
flight-booking/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── app/             # Application setup and DI
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── interfaces/      # Interface definitions
│   ├── models/          # Data models
│   ├── providers/       # External API providers
│   ├── services/        # Shared services
│   └── usecases/        # Business logic
├── api/                 # OpenAPI specifications
├── docs/                # Generated documentation
├── .github/workflows/   # CI/CD pipelines
├── Dockerfile           # Container configuration
├── docker-compose.yml   # Local development setup
├── Taskfile.yml         # Build automation
└── README.md
```

### Available Tasks

```bash
# Development
task dev                 # Run with hot reload
task generate           # Generate code from OpenAPI spec
task build              # Build the application
task run                # Run the application

# Testing
task test               # Run tests
task test-coverage      # Run tests with coverage
task lint               # Run linter
task lint-fix           # Run linter with auto-fix

# Code Generation
task generate-models    # Generate models from OpenAPI
task generate-mocks     # Generate mocks for testing
task generate-swagger   # Generate swagger documentation

# Docker
task docker-build       # Build Docker image
task docker-run         # Run Docker container

# Utilities
task clean              # Clean generated files
task mod-tidy           # Tidy Go modules
```

### Adding New Providers

1. Create a new provider in `internal/providers/`:
```go
type Provider3 struct {
    *BaseProvider
}

func NewProvider3(config config.ProviderConfig, httpClient interfaces.HTTPClient, cache interfaces.CacheService, logger interfaces.Logger) interfaces.RouteProvider {
    baseProvider := NewBaseProvider("provider3", config, httpClient, cache, logger)
    return &Provider3{BaseProvider: baseProvider}
}

func (p *Provider3) GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.FlightRoute, error) {
    // Implementation
}
```

2. Register the provider in `internal/app/app.go`:
```go
fx.Provide(
    // ... existing providers
    NewProvider3,
),
```

3. Update the provider manager to include the new provider.

### Testing

The project includes comprehensive testing with:

- Unit tests for all components
- Integration tests for API endpoints
- Mock providers for testing
- Test coverage reporting

Run tests:
```bash
task test
task test-coverage
```

### Code Quality

The project enforces code quality through:

- **golangci-lint**: Static code analysis
- **mockery**: Automatic mock generation
- **gofumpt**: Code formatting
- **Pre-commit hooks**: Automated quality checks

## Deployment

### Docker Deployment

1. Build the image:
```bash
docker build -t flight-booking:latest .
```

2. Run the container:
```bash
docker run -p 8080:8080 \
  -e FLIGHT_BOOKING_PROVIDERS_PROVIDER1_BASE_URL=https://api.provider1.com \
  -e FLIGHT_BOOKING_PROVIDERS_PROVIDER2_BASE_URL=https://api.provider2.com \
  flight-booking:latest
```

### Kubernetes Deployment

Example Kubernetes deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flight-booking
spec:
  replicas: 3
  selector:
    matchLabels:
      app: flight-booking
  template:
    metadata:
      labels:
        app: flight-booking
    spec:
      containers:
      - name: flight-booking
        image: flight-booking:latest
        ports:
        - containerPort: 8080
        env:
        - name: FLIGHT_BOOKING_PROVIDERS_PROVIDER1_BASE_URL
          value: "https://api.provider1.com"
        - name: FLIGHT_BOOKING_PROVIDERS_PROVIDER2_BASE_URL
          value: "https://api.provider2.com"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Cloud Deployment Options

The application is designed to be cloud-agnostic and can be deployed on:

- **AWS**: ECS, EKS, Lambda (with API Gateway)
- **Google Cloud**: GKE, Cloud Run, Cloud Functions
- **Azure**: AKS, Container Instances, Functions
- **Any Kubernetes cluster**

## Monitoring and Observability

### Logging

The application uses structured logging with configurable levels:

```json
{
  "level": "info",
  "timestamp": "2023-12-01T10:30:00Z",
  "message": "Successfully fetched routes",
  "component": "route_service",
  "total_routes": 150,
  "providers_used": ["provider1", "provider2"],
  "cache_hit": false
}
```

### Health Checks

- **Application Health**: `/health` endpoint
- **Provider Health**: Individual provider status
- **Dependency Health**: Cache and other services

### Metrics

Ready for integration with:
- Prometheus
- Grafana
- DataDog
- New Relic

## Performance Considerations

- **Concurrent Provider Calls**: All providers are called concurrently
- **Intelligent Caching**: Configurable TTL to balance freshness and performance
- **Connection Pooling**: HTTP client with connection reuse
- **Graceful Degradation**: Continues to serve if one provider fails
- **Request Deduplication**: Prevents duplicate requests
- **Circuit Breaker**: Can be added for provider resilience

## Security

- **No Third-Party Dependencies**: Uses only open-source libraries
- **Input Validation**: All inputs are validated
- **Security Headers**: CORS, security headers configured
- **Container Security**: Non-root user in Docker
- **Secrets Management**: Environment variable based configuration

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Run the linter
7. Submit a pull request

### Code Style

- Follow Go best practices
- Use meaningful variable names
- Add comments for public APIs
- Write tests for new functionality
- Keep functions small and focused

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions and support:
- Create an issue in the repository
- Contact the development team
- Check the API documentation

---

**Built with ❤️ for FunWithFlights** 