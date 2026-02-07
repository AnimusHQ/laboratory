export function getGatewayLoginUrl(returnToPath: string = '/console'): string {
  const base = process.env.NEXT_PUBLIC_GATEWAY_URL?.trim();
  if (!base) {
    throw new Error('NEXT_PUBLIC_GATEWAY_URL is required to build login URL');
  }
  const normalized = base.endsWith('/') ? base.slice(0, -1) : base;
  return `${normalized}/auth/login?return_to=${encodeURIComponent(returnToPath)}`;
}
