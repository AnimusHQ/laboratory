import { strict as assert } from 'node:assert';
import { afterEach, test } from 'node:test';

import { getGatewayLoginUrl } from '../lib/auth/login-url';

const originalGateway = process.env.NEXT_PUBLIC_GATEWAY_URL;

afterEach(() => {
  if (originalGateway === undefined) {
    delete process.env.NEXT_PUBLIC_GATEWAY_URL;
  } else {
    process.env.NEXT_PUBLIC_GATEWAY_URL = originalGateway;
  }
});

test('getGatewayLoginUrl builds URL with encoded return_to', () => {
  process.env.NEXT_PUBLIC_GATEWAY_URL = 'http://localhost:8080/';
  const url = getGatewayLoginUrl('/console?x=1&y=space here');
  assert.equal(url, 'http://localhost:8080/auth/login?return_to=%2Fconsole%3Fx%3D1%26y%3Dspace%20here');
});

test('getGatewayLoginUrl throws when gateway URL missing', () => {
  delete process.env.NEXT_PUBLIC_GATEWAY_URL;
  assert.throws(() => getGatewayLoginUrl('/console'), /NEXT_PUBLIC_GATEWAY_URL is required/);
});
