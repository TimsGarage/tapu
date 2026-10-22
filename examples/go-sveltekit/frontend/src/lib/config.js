import { env } from '$env/dynamic/public';

function requiredOrigin(name, value) {
  const normalized = value
    ?.trim()
    .replace(/\/$/, '');

  if (!normalized) {
    throw new Error(
      `Missing required environment variable: ${name}`
    );
  }

  let parsed;

  try {
    parsed = new URL(normalized);
  } catch {
    throw new Error(`${name} must be a valid URL`);
  }

  if (
    parsed.protocol !== 'http:' &&
    parsed.protocol !== 'https:'
  ) {
    throw new Error(
      `${name} must use HTTP or HTTPS`
    );
  }

  if (
    parsed.pathname !== '/' ||
    parsed.search ||
    parsed.hash ||
    parsed.username ||
    parsed.password
  ) {
    throw new Error(
      `${name} must contain only a scheme and host`
    );
  }

  return normalized;
}

export const config = Object.freeze({
  apiOrigin: requiredOrigin(
    'PUBLIC_API_ORIGIN',
    env.PUBLIC_API_ORIGIN
  )
});