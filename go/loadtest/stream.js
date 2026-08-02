import http from 'k6/http';
import { check } from 'k6';

const listeners = Number.parseInt(__ENV.LISTENERS || '10', 10);
const requestTimeout = __ENV.REQUEST_TIMEOUT || '90s';
const baseURL = (__ENV.BASE_URL || 'https://api.example.com/api/v1').replace(/\/$/, '');

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
  for (const name of ['STREAM_ID', 'LISTENER_TOKEN']) {
    if (!__ENV[name]) {
      throw new Error(`${name} is required`);
    }
  }
}

export default function () {
  // This request stays open until the publisher stops and STREAM_IDLE_TIMEOUT
  // expires. responseType=none prevents k6 from retaining audio in memory.
  const response = http.get(`${baseURL}/streams/${__ENV.STREAM_ID}/listen`, {
    headers: {
      Authorization: `Bearer ${__ENV.LISTENER_TOKEN}`,
      Accept: 'audio/mpeg, audio/aac, application/octet-stream',
    },
    responseType: 'none',
    timeout: requestTimeout,
    tags: { endpoint: 'audio-listen' },
  });
  check(response, {
    'listener received HTTP 200': (r) => r.status === 200,
    'listener received an audio content type': (r) =>
      (r.headers['Content-Type'] || '').startsWith('audio/') ||
      r.headers['Content-Type'] === 'application/octet-stream',
  });
}
