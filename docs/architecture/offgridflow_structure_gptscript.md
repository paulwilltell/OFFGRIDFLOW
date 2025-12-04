Repository Structural Map Report
===============================

Overview
--------
This repository implements OffGridFlow, a carbon accounting and emissions management platform. It provides a backend API server, CLI and worker components, AI integration, emissions calculation engines, compliance reporting, authentication, billing, auditing, ingestion of activity data, and a React-based frontend.

Packages and Key Files
----------------------

1. **cmd/**
   - `api/main.go`: Main API server entrypoint. Initializes configuration, tracing, database, mode manager, AI router, stores, auth, billing, and HTTP router.
   - `cli/main.go`: CLI tool entrypoint with commands for demo data ingestion, emissions recalculation, and CSRD report generation (mostly TODO).
   - `worker/main.go`: Background worker entrypoint (stub for scheduled jobs).

2. **internal/ai/**
   - `openai_provider.go`: Implements OpenAI cloud provider integration with HTTP client, request/response handling, error mapping, and metrics.
   - `router.go`: AI router that routes chat requests between cloud and local providers based on connectivity mode, with fallback and retry logic.
   - `stubs.go`: Stub implementations of cloud and local AI providers for testing.
   - `types.go`: Core AI types, interfaces, errors, and request/response models.

3. **internal/allocation/**
   - `rules.go`: Defines allocation rules, dimensions, methods, validation, and a rule set manager.
   - `service.go`: Allocation service applying rules to emissions, managing driver data, metrics, and batch processing.
   - `service_test.go`: Unit tests for allocation service metrics.

4. **internal/api/**
   - `graph/`: GraphQL resolvers and production resolver implementations.
     - `production_resolver.go`: Real resolver using activity store and emissions calculators.
     - `resolvers.go`: Default stub resolvers and GraphQL types.
   - `http/`: HTTP router and handlers.
     - `router.go`: HTTP router setup with public and protected routes, middleware for auth and billing.
     - `handlers/`: HTTP handlers for auth, billing, emissions, compliance, scope2, health, etc.
     - `middleware/`: Auth and subscription enforcement middleware.
     - `responders/`: Standardized JSON response helpers.

5. **internal/audit/**
   - `models.go`: Audit entry data models, builder, and query types.
   - `service.go`: Audit logging service with async buffering, masking, in-memory store, and retention.

6. **internal/auth/**
   - `auth.go`: Role-based authorization (RBAC) implementation.
   - `models.go`: Auth domain models (Tenant, User, APIKey) and context helpers.
   - `password.go`: Password policy, validation, and bcrypt hashing.
   - `service.go`: High-level auth service managing tenants, users, API keys, and authorization.
   - `session.go`: JWT session token management.
   - `store.go`: Persistence interfaces and PostgreSQL/in-memory implementations.

7. **internal/billing/**
   - `service.go`: Billing service integrating with Stripe, subscription management, webhook handling.
   - `stripe_client.go`: Stripe API client wrapper.
   - `subscription_model.go`: Subscription data model and status.

8. **internal/compliance/**
   - Subpackages for different compliance frameworks: `csrd`, `sec`, `cbam`, `california`.
   - Each has mappers and validators (mostly stubs or TODO).
   - `core/`: Core compliance interfaces and templates.

9. **internal/config/**
   - `config.go`: Centralized configuration loading from environment variables with validation.

10. **internal/db/**
    - `db.go`: PostgreSQL connection management, migrations, transaction helpers.

11. **internal/emissions/**
    - Core emissions calculation engine and models.
    - Scope-specific calculators: `scope1.go`, `scope2.go`, `scope3.go`.
    - Factor registries: in-memory and PostgreSQL-backed.
    - Testing helpers and unit tests.

12. **internal/events/**
    - Domain event bus interfaces and in-memory/noop implementations.

13. **internal/ingestion/**
    - Activity data ingestion models and stores.
    - Source adapters for AWS, Azure, GCP, SAP, utility bills, CSV uploads (mostly stubs).

14. **internal/logging/**
    - Structured logging setup using Go slog package.

15. **internal/offgrid/**
    - Connectivity mode management: ModeManager and ConnectivityWatcher for online/offline detection.

16. **internal/reporting/**
    - Export generators for Excel, PDF, XBRL (stubs).

17. **internal/tracing/**
    - OpenTelemetry tracing setup and helpers.

18. **internal/workflow/**
    - Workflow task models and service (stub).

19. **web/**
    - React frontend with Next.js app directory.
    - Pages for dashboard, login, register, billing, emissions, compliance reports, workflow, settings.
    - Layout and UI components.

Main Responsibilities
---------------------

- **cmd/api**: Bootstraps the API server, wiring all components and starting HTTP server.
- **internal/api/http**: HTTP routing, authentication, billing, AI chat, emissions, and compliance API endpoints.
- **internal/api/graph**: GraphQL API resolvers for querying emissions and compliance data.
- **internal/ai**: AI chat providers and routing between cloud and local inference.
- **internal/emissions**: Emissions calculation engine, scope-specific calculators, and factor registries.
- **internal/ingestion**: Data ingestion from various sources into activity store.
- **internal/auth**: Authentication, authorization, session management, API key handling.
- **internal/billing**: Subscription management with Stripe integration.
- **internal/audit**: Audit logging of system actions and data changes.
- **internal/offgrid**: Connectivity mode detection and management for offline support.
- **internal/compliance**: Compliance report generation and validation for various frameworks.
- **internal/config**: Configuration loading and validation.
- **internal/db**: Database connection and migration management.
- **internal/events**: Event bus for domain events.
- **internal/logging**: Structured logging.
- **internal/tracing**: Distributed tracing instrumentation.
- **internal/reporting**: Report export stubs.
- **internal/workflow**: Workflow task management stub.
- **web**: React frontend UI.

Data Flow
---------

1. **Data Ingestion**: Activity data is ingested from various sources (CSV uploads, cloud billing, utility bills) into the `ingestion.ActivityStore` (Postgres or in-memory).

2. **Emissions Calculation**: Activities are passed to the `emissions.Engine` which routes to scope-specific calculators (`scope1`, `scope2`, `scope3`) using emission factors from the `emissions.FactorRegistry` (Postgres or in-memory).

3. **Allocation**: Calculated emissions can be allocated across organizational dimensions using the `allocation.Service` applying configured allocation rules.

4. **Compliance Reporting**: Emissions data is mapped to compliance reports (CSRD, SEC, CBAM, California) via mappers in `compliance` packages.

5. **API Layer**: The API server exposes HTTP REST endpoints and GraphQL queries/mutations to access activities, emissions, compliance reports, billing, and auth.

6. **AI Integration**: AI chat requests are routed through the `ai.Router` which selects cloud (OpenAI) or local providers based on connectivity mode managed by `offgrid.ModeManager` and `ConnectivityWatcher`.

7. **Authentication & Authorization**: Auth middleware validates JWT sessions or API keys, enforcing RBAC permissions and subscription status via billing middleware.

8. **Billing**: Stripe integration manages subscriptions, checkout, and billing portal sessions.

9. **Audit Logging**: All significant actions are logged asynchronously to the audit service.

10. **Frontend**: React app consumes API endpoints to provide UI for dashboard, emissions exploration, compliance reports, billing, and user management.

Obvious Risks or Missing Parts
------------------------------

- **Incomplete Implementations**: Many ingestion adapters (AWS, Azure, GCP, SAP, utility bills) are stubs with TODOs, limiting data source coverage.

- **Compliance Mappers and Validators**: SEC, CBAM, California compliance mappers and validators are mostly stubs or TODO, lacking full regulatory support.

- **Expression-based Allocation**: Allocation expression evaluation is not implemented; currently falls back to equal distribution.

- **AI Local Provider**: The local AI provider is a simple stub; no real local inference engine integrated.

- **Worker Component**: The worker main is a stub with TODOs for scheduled jobs like calculations and syncing.

- **Error Handling and Logging**: Some error paths log warnings but do not propagate errors, which may hide issues.

- **Security**: Default JWT secret fallback in API server is insecure for production; requires environment configuration.

- **Testing Coverage**: While some unit tests exist (allocation, emissions calculators), coverage for API handlers, auth flows, billing, and compliance is not shown.

- **Scalability**: In-memory stores are used as fallback; not suitable for production scale or persistence.

- **Feature Flags**: Some compliance features are gated by feature flags but no flag management system is shown.

- **Frontend**: UI pages use mock data; no integration with real API data shown in frontend code.

Summary
-------

The repository is well-structured with clear separation of concerns across packages. It provides a comprehensive backend for carbon accounting with modular AI integration, emissions calculation, compliance reporting, authentication, billing, and auditing. The API server is robust with middleware for auth and subscription enforcement.

However, several key components are incomplete or stubbed, especially ingestion adapters, compliance mappers, and worker jobs. The local AI provider is minimal, and frontend pages mostly use mock data. Security defaults require attention for production readiness.

To improve, focus should be on completing ingestion connectors, implementing compliance logic, enhancing local AI support, expanding tests, and integrating frontend with real API data. Also, ensure secure configuration management and robust error handling.

This structure supports extensibility and maintainability, enabling OffGridFlow to evolve into a full-featured carbon accounting platform.