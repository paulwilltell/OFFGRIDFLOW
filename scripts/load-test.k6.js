// OffGridFlow Load Testing Suite using k6
// Tests API endpoints under realistic load patterns
// Run: k6 run scripts/load-test.k6.js

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

// Custom metrics
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const activityCreateDuration = new Trend('activity_create_duration');
const reportsGenerateDuration = new Trend('report_generate_duration');
const apiRequestsCounter = new Counter('api_requests_total');

// Test configuration
export const options = {
  stages: [
    // Warm-up
    { duration: '2m', target: 10 },   // Ramp up to 10 users
    
    // Normal load
    { duration: '5m', target: 50 },   // Ramp up to 50 users
    { duration: '10m', target: 50 },  // Stay at 50 users
    
    // Peak load
    { duration: '3m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    
    // Stress test
    { duration: '2m', target: 200 },  // Spike to 200 users
    { duration: '3m', target: 200 },  // Hold spike
    
    // Cool down
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  
  thresholds: {
    // HTTP errors should be less than 1%
    'errors': ['rate<0.01'],
    
    // 95% of requests should be below 500ms
    'http_req_duration': ['p(95)<500'],
    
    // 99% of requests should be below 2s
    'http_req_duration': ['p(99)<2000'],
    
    // Specific endpoint thresholds
    'login_duration': ['p(95)<300'],
    'activity_create_duration': ['p(95)<400'],
    'report_generate_duration': ['p(95)<2000'],
  },
};

// Configuration
const BASE_URL = __ENV.API_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY || 'test-api-key';

// Test data
const TEST_USERS = [
  { email: 'loadtest1@example.com', password: 'LoadTest123!' },
  { email: 'loadtest2@example.com', password: 'LoadTest123!' },
  { email: 'loadtest3@example.com', password: 'LoadTest123!' },
  { email: 'loadtest4@example.com', password: 'LoadTest123!' },
  { email: 'loadtest5@example.com', password: 'LoadTest123!' },
];

// Helper function to get random test user
function getRandomUser() {
  return TEST_USERS[Math.floor(Math.random() * TEST_USERS.length)];
}

// Setup function - runs once before test
export function setup() {
  console.log(`Starting load test against ${BASE_URL}`);
  
  // Verify API is reachable
  const healthCheck = http.get(`${BASE_URL}/health`);
  if (healthCheck.status !== 200) {
    throw new Error(`API health check failed: ${healthCheck.status}`);
  }
  
  console.log('API is healthy, starting load test...');
  return { baseUrl: BASE_URL };
}

// Main test scenario
export default function(data) {
  const user = getRandomUser();
  let authToken = null;
  
  // 1. User Login
  group('Authentication', () => {
    const loginStart = Date.now();
    const loginRes = http.post(`${data.baseUrl}/v1/auth/login`, 
      JSON.stringify({
        email: user.email,
        password: user.password,
      }),
      {
        headers: {
          'Content-Type': 'application/json',
        },
      }
    );
    
    const success = check(loginRes, {
      'login status is 200': (r) => r.status === 200,
      'login has token': (r) => r.json('token') !== undefined,
      'login response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    if (success && loginRes.status === 200) {
      authToken = loginRes.json('token');
      loginDuration.add(Date.now() - loginStart);
    } else {
      errorRate.add(1);
      console.error(`Login failed: ${loginRes.status}`);
      return;
    }
    
    apiRequestsCounter.add(1);
    sleep(1);
  });
  
  // Only proceed if login succeeded
  if (!authToken) {
    return;
  }
  
  const headers = {
    'Authorization': `Bearer ${authToken}`,
    'Content-Type': 'application/json',
  };
  
  // 2. Get Dashboard Data
  group('Dashboard', () => {
    const dashboardRes = http.get(`${data.baseUrl}/v1/dashboard`, {
      headers: headers,
    });
    
    check(dashboardRes, {
      'dashboard status is 200': (r) => r.status === 200,
      'dashboard has data': (r) => r.json('data') !== undefined,
      'dashboard response time < 300ms': (r) => r.timings.duration < 300,
    }) || errorRate.add(1);
    
    apiRequestsCounter.add(1);
    sleep(1);
  });
  
  // 3. List Activities
  group('Activities - List', () => {
    const activitiesRes = http.get(`${data.baseUrl}/v1/activities?limit=20`, {
      headers: headers,
    });
    
    check(activitiesRes, {
      'activities list status is 200': (r) => r.status === 200,
      'activities list has items': (r) => r.json('activities') !== undefined,
      'activities response time < 400ms': (r) => r.timings.duration < 400,
    }) || errorRate.add(1);
    
    apiRequestsCounter.add(1);
    sleep(2);
  });
  
  // 4. Create Activity (Write Operation)
  group('Activities - Create', () => {
    const createStart = Date.now();
    const activityData = {
      name: `Load Test Activity ${Date.now()}`,
      activity_type: 'electricity',
      scope: 2,
      quantity: Math.random() * 1000,
      unit: 'kWh',
      activity_date: new Date().toISOString().split('T')[0],
    };
    
    const createRes = http.post(
      `${data.baseUrl}/v1/activities`,
      JSON.stringify(activityData),
      { headers: headers }
    );
    
    const success = check(createRes, {
      'create activity status is 201': (r) => r.status === 201,
      'create activity has id': (r) => r.json('id') !== undefined,
      'create response time < 600ms': (r) => r.timings.duration < 600,
    });
    
    if (success) {
      activityCreateDuration.add(Date.now() - createStart);
    } else {
      errorRate.add(1);
    }
    
    apiRequestsCounter.add(1);
    sleep(2);
  });
  
  // 5. Get Emission Factors
  group('Emission Factors', () => {
    const factorsRes = http.get(`${data.baseUrl}/v1/emission-factors?limit=50`, {
      headers: headers,
    });
    
    check(factorsRes, {
      'emission factors status is 200': (r) => r.status === 200,
      'emission factors response time < 500ms': (r) => r.timings.duration < 500,
    }) || errorRate.add(1);
    
    apiRequestsCounter.add(1);
    sleep(1);
  });
  
  // 6. Generate Report (Heavy Operation)
  group('Reports - Generate', () => {
    const reportStart = Date.now();
    const reportData = {
      report_type: 'csrd',
      report_year: 2024,
      reporting_period_start: '2024-01-01',
      reporting_period_end: '2024-12-31',
    };
    
    const reportRes = http.post(
      `${data.baseUrl}/v1/reports/generate`,
      JSON.stringify(reportData),
      { headers: headers }
    );
    
    const success = check(reportRes, {
      'report generation status is 202': (r) => r.status === 202 || r.status === 200,
      'report response time < 3000ms': (r) => r.timings.duration < 3000,
    });
    
    if (success) {
      reportsGenerateDuration.add(Date.now() - reportStart);
    } else {
      errorRate.add(1);
    }
    
    apiRequestsCounter.add(1);
    sleep(3);
  });
  
  // 7. Get Compliance Status
  group('Compliance', () => {
    const complianceRes = http.get(`${data.baseUrl}/v1/compliance/status`, {
      headers: headers,
    });
    
    check(complianceRes, {
      'compliance status is 200': (r) => r.status === 200,
      'compliance response time < 400ms': (r) => r.timings.duration < 400,
    }) || errorRate.add(1);
    
    apiRequestsCounter.add(1);
    sleep(2);
  });
  
  // Random think time between iterations
  sleep(Math.random() * 3 + 1);
}

// Teardown function - runs once after test
export function teardown(data) {
  console.log('Load test completed');
}

// Generate HTML report
export function handleSummary(data) {
  return {
    "load-test-results.html": htmlReport(data),
    "load-test-results.json": JSON.stringify(data),
  };
}
