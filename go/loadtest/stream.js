import http from 'k6/http';
import { check, sleep } from 'k6';

const listeners = Number.parseInt(__ENV.LISTENERS || '10', 10);
const startRampSeconds = Number.parseFloat(__ENV.START_RAMP_SECONDS || '10');
const requestTimeout = __ENV.REQUEST_TIMEOUT || '90s';
const baseURL = (__ENV.BASE_URL || 'https://api.example.com/api/v1').replace(/\/$/, '');
const listenerToken = __ENV.LISTENER_TOKEN || '';
const listenerPath = __ENV.LISTENER_PATH || (listenerToken ? 'listen' : 'audio');
const resultsDirectory = (__ENV.RESULTS_DIR || 'loadtest/results').replace(/\/$/, '');
const resultPrefix = __ENV.RESULT_PREFIX || `k6-${listeners}`;

if (!Number.isInteger(listeners) || listeners < 1) {
  throw new Error('LISTENERS must be a positive integer');
}
if (!Number.isFinite(startRampSeconds) || startRampSeconds < 0) {
  throw new Error('START_RAMP_SECONDS must be a non-negative number');
}
if (!['audio', 'listen'].includes(listenerPath)) {
  throw new Error('LISTENER_PATH must be audio or listen');
}

export const options = {
  scenarios: {
    audio_listeners: {
      executor: 'per-vu-iterations',
      vus: listeners,
      iterations: 1,
      maxDuration: __ENV.MAX_DURATION || '2m',
      gracefulStop: '5s',
    },
  },
  thresholds: {
    checks: ['rate>0.99'],
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<95000'],
  },
  discardResponseBodies: true,
};

export function setup() {
  for (const name of ['STREAM_ID']) {
    if (!__ENV[name]) {
      throw new Error(`${name} is required`);
    }
  }
  if (listenerPath === 'listen' && !listenerToken) {
    throw new Error('LISTENER_TOKEN is required when LISTENER_PATH=listen');
  }
}

export default function () {
  // Spread only connection establishment. Once started, every request remains
  // open, so the requested listener tier is concurrently active without
  // overflowing a local TCP accept backlog in a single instant.
  if (startRampSeconds > 0 && listeners > 1) {
    sleep(((__VU - 1) / (listeners - 1)) * startRampSeconds);
  }

  // This request stays open until the publisher stops and STREAM_IDLE_TIMEOUT
  // expires. responseType=none prevents k6 from retaining audio in memory.
  const headers = {
    Accept: 'audio/mpeg, audio/aac, application/octet-stream',
  };
  if (listenerToken) {
    headers.Authorization = `Bearer ${listenerToken}`;
  }
  const response = http.get(`${baseURL}/streams/${__ENV.STREAM_ID}/${listenerPath}`, {
    headers,
    responseType: 'none',
    timeout: requestTimeout,
    tags: { endpoint: `audio-${listenerPath}`, tier: String(listeners) },
  });
  check(response, {
    'listener received HTTP 200': (r) => r.status === 200,
    'listener received an audio content type': (r) =>
      (r.headers['Content-Type'] || '').startsWith('audio/') ||
      r.headers['Content-Type'] === 'application/octet-stream',
  });
}

export function handleSummary(data) {
  return {
    [`${resultsDirectory}/${resultPrefix}-summary.json`]: JSON.stringify(data, null, 2),
  };
}
