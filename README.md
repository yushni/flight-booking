# Flight Booking Routes API

> **Important Note**: Implementation includes only the backend API implementation. No UI components are included.

A cloud-native flight booking aggregator system that provides a unified API for flight route information from multiple providers.

**Live API**: http://flight-booking-lb-1909835177.eu-central-1.elb.amazonaws.com/api/v1/routes

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │───▶│      ECS        │───▶│  Flight Booking │
│   (AWS ALB)     │    │   (Fargate)     │    │     Service     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

The application is deployed on AWS using:
- **Load Balancer**: AWS Application Load Balancer
- **Container Orchestration**: Amazon ECS with Fargate
- **Container**: Docker containerized Go application

## CI/CD

The project includes automated CI/CD pipeline that:
- Runs tests and linting on every commit
- Builds and deploys the application when a release is created
- Automatically updates the ECS service with the new container image

## Development

### Prerequisites
- Go 1.24 or higher
- Task (optional, for build automation)

### Quick Start

1. Clone and install dependencies:
```bash
git clone <repository-url>
cd flight-booking
go mod download
```

2. Generate code and run linter:
```bash
task generate  # Generate models from OpenAPI spec
task lint      # Run linter with auto-fix
```

3. Run the application:
```bash
go run main.go
```

### Available Tasks

Run `task` in this directory to see all available tasks with descriptions.

## API Endpoints

The API follows the OpenAPI 3.0 specification defined in `openapi.yaml`. This project uses an API-first approach where the schema is defined first and then code is generated from it.

## Configuration

### Environment Variables

To configure the deployment, modify the environment variables in `infra/ecs/ecs-task-definition.json`:

```json
{
  "environment": [
    {
      "name": "PROVIDER1_CACHE_TTL",
      "value": "3600s"
    },
    {
      "name": "PROVIDER2_CACHE_TTL", 
      "value": "3600s"
    }
  ]
}
```

## Features

- **Multi-Provider Aggregation**: Fetches flight routes from multiple providers
- **Intelligent Caching**: Configurable TTL-based caching
- **Clean Architecture**: Separated concerns with handlers, use cases, and providers
- **Production Ready**: Logging, monitoring, graceful shutdown, and error handling 