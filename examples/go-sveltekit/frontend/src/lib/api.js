import { config } from './config.js';

export class APIError extends Error {
  constructor(status, message) {
    super(message);
    this.name = 'APIError';
    this.status = status;
  }
}

export async function apiRequest(
  path,
  options = {}
) {
  const headers = new Headers(
    options.headers || {}
  );

  headers.set('Accept', 'application/json');

  if (options.body !== undefined) {
    headers.set(
      'Content-Type',
      'application/json'
    );
  }

  const response = await fetch(
    `${config.apiOrigin}${path}`,
    {
      ...options,
      credentials: 'include',
      headers,
      body:
        options.body === undefined
          ? undefined
          : JSON.stringify(options.body)
    }
  );

  if (response.status === 204) {
    return null;
  }

  const body = await response
    .json()
    .catch(() => ({}));

  if (!response.ok) {
    throw new APIError(
      response.status,
      body.error || 'Request failed'
    );
  }

  return body;
}