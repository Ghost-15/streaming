import http from 'k6/http';
import { check, sleep } from 'k6';

const baseURL = (__ENV.BASE_URL || 'http://127.0.0.1:8080/api/v1').replace(/\/$/, '');
const publishSeconds = Number.parseInt(__ENV.PUBLISH_SECONDS || '60', 10);
const intervalSeconds = Number.parseFloat(__ENV.PUBLISH_INTERVAL || '0.5');
const chunkBytes = Number.parseInt(__ENV.PUBLISH_CHUNK_BYTES || '8192', 10);
const resultsDirectory = (__ENV.RESULTS_DIR || 'loadtest/results').replace(/\/$/, '');
const resultPrefix = __ENV.RESULT_PREFIX || 'publisher';

if (!Number.isInteger(publishSeconds) || publishSeconds < 1) {
  throw new Error('PUBLISH_SECONDS must be a positive integer');
}
if (!(intervalSeconds > 0) || !Number.isInteger(chunkBytes) || chunkBytes < 1) {
  throw new Error('PUBLISH_INTERVAL and PUBLISH_CHUNK_BYTES must be positive');
}

export const options = {
  scenarios: {
    audio_publisher: {
      executor: 'constant-vus',
      vus: 1,
      duration: `${publishSeconds}s`,
      gracefulStop: '2s',
    },
  },
  thresholds: {
    checks: ['rate>0.99'],
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<5000'],
  },
  discardResponseBodies: true,
};

export function setup() {
  for (const name of ['STREAM_ID', 'BROADCASTER_TOKEN', 'STREAM_SESSION_ID']) {
    if (!__ENV[name]) {
      throw new Error(`${name} is required`);
    }
  }
}

const payload = new Uint8Array(chunkBytes);
payload[0] = 0xff;
payload[1] = 0xfb;

export default function () {
  const iterationStartedAt = Date.now();
  const response = http.post(
    `${baseURL}/streams/${__ENV.STREAM_ID}/push`,
    payload.buffer,
    {
      headers: {
        Authorization: `Bearer ${__ENV.BROADCASTER_TOKEN}`,
        'X-Stream-Session-ID': __ENV.STREAM_SESSION_ID,
        'Content-Type': 'audio/mpeg',
      },
      timeout: '10s',
      tags: { endpoint: 'audio-push' },
    },
  );
  check(response, {
    'publisher received HTTP 204': (r) => r.status === 204,
  });
  // Pace start-to-start, not response-to-start: HTTP handling time must not be
  // added to the configured audio interval. 8192 bytes every 0.5 s is
  // 131072 bit/s (128 Kibit/s).
  const remainingSeconds = intervalSeconds - (Date.now() - iterationStartedAt) / 1000;
  if (remainingSeconds > 0) {
    sleep(remainingSeconds);
  }
}

export function handleSummary(data) {
  const measuredSeconds = data.state.testRunDurationMs / 1000;
  data.publisher = {
    publish_seconds: publishSeconds,
    interval_seconds: intervalSeconds,
    chunk_bytes: chunkBytes,
    configured_payload_kbit_s: (chunkBytes * 8) / intervalSeconds / 1000,
    measured_payload_kbit_s:
      (data.metrics.iterations.values.count * chunkBytes * 8) / measuredSeconds / 1000,
  };
  return {
    [`${resultsDirectory}/${resultPrefix}-summary.json`]: JSON.stringify(data, null, 2),
  };
}
