# OffGridFlow GraphQL API Reference

Complete GraphQL API documentation for OffGridFlow carbon accounting platform.

## Endpoint

```
Production: https://api.offgridflow.com/graphql
Development: http://localhost:8080/graphql
```

## Authentication

Include your JWT token in the Authorization header:

```http
Authorization: Bearer <access_token>
```

---

## Schema Overview

```graphql
type Query {
  # Emissions data
  emissions(scope: Int, startDate: String, endDate: String): [Emission!]!
  emissionsSummary(startDate: String, endDate: String): EmissionsSummary!
  
  # Compliance
  complianceStatus(framework: Framework): ComplianceStatus!
  frameworks: [FrameworkInfo!]!
  
  # Activities
  activities(first: Int, after: String, filter: ActivityFilter): ActivityConnection!
  activity(id: ID!): Activity
  
  # User & Tenant
  me: User!
  tenant: Tenant!
  tenants: [Tenant!]!
  
  # AI & Analytics
  anomalies(severity: Severity): [Anomaly!]!
  benchmarks(industry: Industry): BenchmarkResult!
  scenarios: [Scenario!]!
}

type Mutation {
  # Activities
  createActivity(input: CreateActivityInput!): Activity!
  updateActivity(id: ID!, input: UpdateActivityInput!): Activity!
  deleteActivity(id: ID!): Boolean!
  
  # Emissions
  createEmission(input: CreateEmissionInput!): Emission!
  updateEmission(id: ID!, input: UpdateEmissionInput!): Emission!
  deleteEmission(id: ID!): Boolean!
  
  # Compliance
  generateReport(framework: Framework!, input: ReportInput!): Report!
  submitEvidence(requirementId: ID!, input: EvidenceInput!): Evidence!
  
  # AI
  chat(prompt: String!, context: ChatContext): ChatResponse!
  generateNarrative(type: NarrativeType!, dataRange: DateRangeInput!): Narrative!
  
  # Scenarios
  createScenario(input: ScenarioInput!): Scenario!
  runScenario(id: ID!): ScenarioResult!
  
  # User management
  updateUser(input: UpdateUserInput!): User!
  enable2FA: TwoFactorSetup!
  disable2FA(otp: String!): Boolean!
  
  # Tenant
  switchTenant(tenantId: ID!): AuthPayload!
}

type Subscription {
  # Real-time emissions updates
  emissionCreated: Emission!
  emissionUpdated: Emission!
  
  # Anomaly alerts
  anomalyDetected: Anomaly!
  
  # Compliance deadlines
  complianceDeadlineApproaching(daysAhead: Int!): ComplianceAlert!
  
  # Sync status
  syncStatusChanged: SyncStatus!
}
```

---

## Types

### Emission

```graphql
type Emission {
  id: ID!
  tenantId: ID!
  scope: Int!
  category: String!
  sourceType: String!
  description: String
  quantity: Float!
  unit: String!
  emissionsTCO2e: Float!
  emissionFactor: Float!
  calculationMethod: String!
  region: String
  periodStart: String!
  periodEnd: String!
  dataQualityScore: Float
  createdAt: String!
  updatedAt: String!
  createdBy: User
}
```

### EmissionsSummary

```graphql
type EmissionsSummary {
  totalEmissionsTCO2e: Float!
  scope1: ScopeSummary!
  scope2: ScopeSummary!
  scope3: ScopeSummary!
  period: DateRange!
  trends: EmissionsTrend
  comparisonToPrevious: Float
}

type ScopeSummary {
  total: Float!
  categories: [CategorySummary!]!
  percentageOfTotal: Float!
}

type CategorySummary {
  name: String!
  emissions: Float!
  percentage: Float!
  activityCount: Int!
}
```

### Activity

```graphql
type Activity {
  id: ID!
  tenantId: ID!
  type: String!
  scope: Int!
  description: String!
  quantity: Float!
  unit: String!
  emissionsTCO2e: Float
  metadata: JSON
  periodStart: String!
  periodEnd: String!
  supplier: Supplier
  facility: Facility
  createdAt: String!
  updatedAt: String!
}

type ActivityConnection {
  edges: [ActivityEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type ActivityEdge {
  node: Activity!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}
```

### Compliance

```graphql
type ComplianceStatus {
  framework: Framework!
  frameworkName: String!
  status: ComplianceState!
  completionPercentage: Float!
  requirements: [Requirement!]!
  deadline: String
  lastUpdated: String!
  missingItems: [String!]!
}

type Requirement {
  id: ID!
  name: String!
  description: String!
  status: RequirementStatus!
  evidence: [Evidence!]!
  dueDate: String
}

type Evidence {
  id: ID!
  type: String!
  description: String!
  fileUrl: String
  uploadedBy: User!
  uploadedAt: String!
  verified: Boolean!
}

enum Framework {
  CSRD
  SEC
  CBAM
  IFRS_S2
  CALIFORNIA
}

enum ComplianceState {
  NOT_STARTED
  IN_PROGRESS
  COMPLETE
  PARTIAL
}

enum RequirementStatus {
  PENDING
  IN_PROGRESS
  COMPLETE
  NOT_APPLICABLE
}
```

### User & Tenant

```graphql
type User {
  id: ID!
  email: String!
  name: String!
  role: Role!
  defaultTenantId: ID
  tenants: [Tenant!]!
  twoFactorEnabled: Boolean!
  createdAt: String!
}

type Tenant {
  id: ID!
  name: String!
  industry: Industry!
  settings: TenantSettings!
  users: [User!]!
  createdAt: String!
}

type TenantSettings {
  dataResidency: Region!
  defaultCurrency: String!
  fiscalYearStart: Int!
  baselineYear: Int
}

enum Role {
  VIEWER
  CONTRIBUTOR
  MANAGER
  ADMIN
  SUPER_ADMIN
}

enum Industry {
  MANUFACTURING
  TECHNOLOGY
  RETAIL
  TRANSPORT
  ENERGY
  FINANCE
  HEALTHCARE
  OTHER
}

enum Region {
  US
  EU
  UK
  APAC
}
```

### AI & Analytics

```graphql
type Anomaly {
  id: ID!
  type: AnomalyType!
  severity: Severity!
  description: String!
  dataPoint: DataPoint!
  expected: Float!
  actual: Float!
  deviation: Float!
  confidence: Float!
  detectedAt: String!
  status: AnomalyStatus!
  resolvedAt: String
  resolvedBy: User
}

enum AnomalyType {
  SPIKE
  DROP
  PATTERN
  MISSING
  OUTLIER
}

enum Severity {
  LOW
  MEDIUM
  HIGH
  CRITICAL
}

enum AnomalyStatus {
  NEW
  INVESTIGATING
  RESOLVED
  FALSE_POSITIVE
}

type BenchmarkResult {
  tenantId: ID!
  industry: Industry!
  percentile: Int!
  peerAverage: Float!
  topQuartile: Float!
  yourValue: Float!
  trend: String!
  recommendations: [String!]!
  calculatedAt: String!
}

type ChatResponse {
  response: String!
  dataReferences: [DataReference!]
  suggestedActions: [String!]
  charts: [ChartData!]
}

type Narrative {
  id: ID!
  type: NarrativeType!
  content: String!
  dataRange: DateRange!
  confidence: Float!
  createdAt: String!
}

enum NarrativeType {
  EXECUTIVE_SUMMARY
  TREND_ANALYSIS
  COMPLIANCE_OVERVIEW
  RECOMMENDATIONS
}
```

### Scenarios

```graphql
type Scenario {
  id: ID!
  name: String!
  type: ScenarioType!
  description: String
  assumptions: [Assumption!]!
  projections: [Projection!]
  status: ScenarioStatus!
  createdBy: User!
  createdAt: String!
}

type Assumption {
  parameter: String!
  baseValue: Float!
  targetValue: Float!
  changeRate: Float!
  startYear: Int!
  endYear: Int!
}

type Projection {
  year: Int!
  scope1: Float!
  scope2: Float!
  scope3: Float!
  total: Float!
  vsBaseline: Float!
  confidence: Float!
}

enum ScenarioType {
  REDUCTION
  GROWTH
  TRANSITION
  NET_ZERO
}

enum ScenarioStatus {
  DRAFT
  CALCULATED
  APPROVED
  ARCHIVED
}
```

---

## Input Types

### Activity Inputs

```graphql
input CreateActivityInput {
  type: String!
  scope: Int!
  description: String!
  quantity: Float!
  unit: String!
  periodStart: String!
  periodEnd: String!
  supplierId: ID
  facilityId: ID
  metadata: JSON
}

input UpdateActivityInput {
  type: String
  description: String
  quantity: Float
  unit: String
  periodStart: String
  periodEnd: String
  metadata: JSON
}

input ActivityFilter {
  scope: Int
  type: String
  startDate: String
  endDate: String
  supplierId: ID
}
```

### Emission Inputs

```graphql
input CreateEmissionInput {
  scope: Int!
  category: String!
  sourceType: String!
  description: String
  quantity: Float!
  unit: String!
  calculationMethod: String!
  region: String
  periodStart: String!
  periodEnd: String!
}

input UpdateEmissionInput {
  description: String
  quantity: Float
  unit: String
  calculationMethod: String
  region: String
  periodStart: String
  periodEnd: String
}
```

### Report & Evidence Inputs

```graphql
input ReportInput {
  periodStart: String!
  periodEnd: String!
  format: ReportFormat!
  includeNarratives: Boolean
  sections: [String!]
}

input EvidenceInput {
  type: String!
  description: String!
  fileUrl: String
  metadata: JSON
}

enum ReportFormat {
  PDF
  EXCEL
  JSON
  HTML
}
```

### Scenario Inputs

```graphql
input ScenarioInput {
  name: String!
  type: ScenarioType!
  description: String
  assumptions: [AssumptionInput!]!
}

input AssumptionInput {
  parameter: String!
  baseValue: Float!
  targetValue: Float!
  changeRate: Float!
  startYear: Int!
  endYear: Int!
}
```

### Date Range

```graphql
input DateRangeInput {
  start: String!
  end: String!
}

type DateRange {
  start: String!
  end: String!
}
```

---

## Query Examples

### Get Emissions Summary

```graphql
query GetEmissionsSummary($startDate: String!, $endDate: String!) {
  emissionsSummary(startDate: $startDate, endDate: $endDate) {
    totalEmissionsTCO2e
    scope1 {
      total
      categories {
        name
        emissions
        percentage
      }
    }
    scope2 {
      total
      categories {
        name
        emissions
        percentage
      }
    }
    scope3 {
      total
      categories {
        name
        emissions
        percentage
      }
    }
    comparisonToPrevious
  }
}
```

**Variables:**
```json
{
  "startDate": "2024-01-01",
  "endDate": "2024-12-31"
}
```

### Get Compliance Status

```graphql
query GetComplianceStatus($framework: Framework!) {
  complianceStatus(framework: $framework) {
    framework
    frameworkName
    status
    completionPercentage
    deadline
    requirements {
      id
      name
      status
      evidence {
        id
        type
        verified
      }
    }
    missingItems
  }
}
```

### List Activities with Pagination

```graphql
query ListActivities($first: Int!, $after: String, $filter: ActivityFilter) {
  activities(first: $first, after: $after, filter: $filter) {
    edges {
      node {
        id
        type
        scope
        description
        quantity
        unit
        emissionsTCO2e
        periodStart
        periodEnd
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

### Get Benchmarks

```graphql
query GetBenchmarks($industry: Industry!) {
  benchmarks(industry: $industry) {
    percentile
    peerAverage
    topQuartile
    yourValue
    trend
    recommendations
    calculatedAt
  }
}
```

---

## Mutation Examples

### Create Activity

```graphql
mutation CreateActivity($input: CreateActivityInput!) {
  createActivity(input: $input) {
    id
    type
    scope
    description
    quantity
    unit
    emissionsTCO2e
    createdAt
  }
}
```

**Variables:**
```json
{
  "input": {
    "type": "electricity_consumption",
    "scope": 2,
    "description": "Main office Q1 electricity",
    "quantity": 25000,
    "unit": "kWh",
    "periodStart": "2024-01-01",
    "periodEnd": "2024-03-31"
  }
}
```

### Generate Report

```graphql
mutation GenerateReport($framework: Framework!, $input: ReportInput!) {
  generateReport(framework: $framework, input: $input) {
    id
    status
    format
    estimatedCompletion
    downloadUrl
  }
}
```

### AI Chat

```graphql
mutation Chat($prompt: String!, $context: ChatContext) {
  chat(prompt: $prompt, context: $context) {
    response
    dataReferences {
      type
      value
      period
    }
    suggestedActions
  }
}
```

### Create Scenario

```graphql
mutation CreateScenario($input: ScenarioInput!) {
  createScenario(input: $input) {
    id
    name
    type
    status
    assumptions {
      parameter
      baseValue
      targetValue
    }
  }
}
```

---

## Subscription Examples

### Real-time Emissions

```graphql
subscription OnEmissionCreated {
  emissionCreated {
    id
    scope
    category
    emissionsTCO2e
    createdAt
    createdBy {
      name
    }
  }
}
```

### Anomaly Alerts

```graphql
subscription OnAnomalyDetected {
  anomalyDetected {
    id
    type
    severity
    description
    expected
    actual
    deviation
    detectedAt
  }
}
```

### Compliance Deadlines

```graphql
subscription OnComplianceDeadline($daysAhead: Int!) {
  complianceDeadlineApproaching(daysAhead: $daysAhead) {
    framework
    requirement
    deadline
    daysRemaining
    status
  }
}
```

---

## Error Handling

GraphQL errors are returned in the standard format:

```json
{
  "data": null,
  "errors": [
    {
      "message": "Unauthorized access",
      "locations": [{"line": 2, "column": 3}],
      "path": ["emissionsSummary"],
      "extensions": {
        "code": "UNAUTHORIZED",
        "timestamp": "2024-06-15T14:30:00Z"
      }
    }
  ]
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `UNAUTHORIZED` | Authentication required or token expired |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `VALIDATION_ERROR` | Invalid input data |
| `RATE_LIMITED` | Too many requests |
| `INTERNAL_ERROR` | Server error |

---

## Best Practices

1. **Use fragments** for reusable field sets
2. **Request only needed fields** to reduce payload size
3. **Use variables** for dynamic values
4. **Handle errors** at both query and field levels
5. **Implement pagination** for large datasets
6. **Use subscriptions** for real-time features

---

*GraphQL API Version: v1*
*Last Updated: 2024*
