import { browser } from '$app/environment';
import { goto } from '$app/navigation';

import { apiRequest, APIError } from './api.js';
import { config } from './config.js';

class AuthenticationState {
  loading = $state(true);
  initialized = $state(false);
  user = $state(null);
  csrfToken = $state('');
  error = $state('');

  authenticated = $derived(Boolean(this.user));

  username = $derived(
    this.user?.username ||
      this.user?.subject ||
      ''
  );

  async initialize() {
    if (!browser || this.initialized) {
      return;
    }

    this.loading = true;
    this.error = '';

    try {
      const result = await apiRequest(
        '/api/auth/me'
      );

      this.user = result.user;
      this.csrfToken = result.csrfToken;
    } catch (error) {
      if (
        error instanceof APIError &&
        error.status === 401
      ) {
        this.user = null;
        this.csrfToken = '';
      } else {
        this.error = messageFrom(error);
      }
    } finally {
      this.initialized = true;
      this.loading = false;
    }
  }

  login() {
    if (!browser) {
      return;
    }

    window.location.assign(
      `${config.apiOrigin}/auth/login`
    );
  }

  async logout() {
    this.loading = true;
    this.error = '';

    try {
      await apiRequest(
        '/api/auth/logout',
        {
          method: 'POST',
          headers: {
            'X-CSRF-Token': this.csrfToken
          }
        }
      );

      this.user = null;
      this.csrfToken = '';

      await goto('/signed-out');
    } catch (error) {
      this.error = messageFrom(error);
    } finally {
      this.loading = false;
    }
  }
}

function messageFrom(error) {
  return error instanceof Error
    ? error.message
    : 'An unexpected error occurred.';
}

export const auth = new AuthenticationState();