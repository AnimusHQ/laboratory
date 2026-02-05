import type { components, paths } from '@/lib/gateway-openapi';

export type ErrorResponse = components['schemas']['ErrorResponse'];

export class GatewayAPIError extends Error {
  status: number;
  code: string;
  requestId?: string;
  details?: unknown;

  constructor(status: number, code: string, requestId?: string, details?: unknown) {
    super(code);
    this.status = status;
    this.code = code;
    this.requestId = requestId;
    this.details = details;
  }
}

export type GatewayFetchOptions = RequestInit & {
  retry?: boolean;
  requestId?: string;
};

export const gatewayBaseURL = (): string => {
  const envBase = process.env.NEXT_PUBLIC_GATEWAY_URL?.trim();
  if (envBase) {
    return envBase.replace(/\/$/, '');
  }
  return '';
};

const isSafeMethod = (method?: string) => {
  if (!method) {
    return true;
  }
  return ['GET', 'HEAD'].includes(method.toUpperCase());
};

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export async function gatewayFetch(path: string, init: GatewayFetchOptions = {}): Promise<Response> {
  const base = gatewayBaseURL();
  const url = base ? `${base}${path}` : path;
  const headers = new Headers(init.headers ?? {});
  if (init.requestId) {
    headers.set('X-Request-Id', init.requestId);
  }

  const attempt = async (): Promise<Response> =>
    fetch(url, {
      ...init,
      headers,
    });

  const res = await attempt();
  if (!res.ok && init.retry !== false && isSafeMethod(init.method) && res.status >= 500) {
    await sleep(350);
    return attempt();
  }
  return res;
}

export async function gatewayFetchJSON<T>(path: string, init: GatewayFetchOptions = {}): Promise<T> {
  const res = await gatewayFetch(path, init);
  if (res.status === 204) {
    return undefined as T;
  }
  const requestId = res.headers.get('X-Request-Id') ?? undefined;
  const contentType = res.headers.get('Content-Type') ?? '';
  if (res.ok) {
    if (!contentType.includes('application/json')) {
      return undefined as T;
    }
    return (await res.json()) as T;
  }

  let parsed: unknown = undefined;
  if (contentType.includes('application/json')) {
    try {
      parsed = await res.json();
    } catch {
      parsed = undefined;
    }
  }
  const errorPayload = parsed as ErrorResponse | undefined;
  const code = errorPayload?.error ?? 'gateway_error';
  const requestIdFinal = errorPayload?.request_id ?? requestId;
  throw new GatewayAPIError(res.status, code, requestIdFinal, parsed);
}

export type GatewayPaths = paths;
