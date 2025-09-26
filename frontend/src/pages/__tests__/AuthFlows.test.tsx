import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import App from '../../App';
import { ToastProvider } from '../../components/ToastProvider';
import { AppStateProvider } from '../../state/AppStateProvider';

type JsonValue = Record<string, unknown> | null;

function createJsonResponse(status: number, body: JsonValue): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    text: async () => (body === null ? '' : JSON.stringify(body))
  } as Response;
}

function extractUrl(input: RequestInfo | URL): string {
  if (typeof input === 'string') {
    return input;
  }
  if (input instanceof URL) {
    return input.toString();
  }
  return input.url;
}

function renderApp(initialEntries: string[]) {
  return render(
    <ToastProvider>
      <AppStateProvider>
        <MemoryRouter initialEntries={initialEntries}>
          <App />
        </MemoryRouter>
      </AppStateProvider>
    </ToastProvider>
  );
}

describe('VidFriends authentication journeys', () => {
  const fetchMock = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>();

  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock);
    fetchMock.mockReset();
    window.localStorage.clear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    window.localStorage.clear();
  });

  it('signs in a user and routes them to the dashboard', async () => {
    const tokens = {
      accessToken: 'access-token',
      accessExpiresAt: new Date(Date.now() + 600_000).toISOString(),
      refreshToken: 'refresh-token',
      refreshExpiresAt: new Date(Date.now() + 86_400_000).toISOString()
    };

    fetchMock.mockImplementation((input, init) => {
      const url = extractUrl(input);
      if (url.includes('/api/v1/auth/login')) {
        return Promise.resolve(createJsonResponse(200, { tokens }));
      }
      if (url.includes('/api/v1/friends')) {
        return Promise.resolve(createJsonResponse(200, { requests: [] }));
      }
      if (url.includes('/api/v1/videos/feed')) {
        return Promise.resolve(createJsonResponse(200, { entries: [] }));
      }
      return Promise.reject(new Error(`Unhandled request: ${url}`));
    });

    const user = userEvent.setup();
    renderApp(['/login']);

    await user.type(screen.getByLabelText(/email/i), 'Alex@Example.com');
    await user.type(screen.getByLabelText(/password/i), 'sup3r-secure');
    await user.click(screen.getByRole('button', { name: /log in/i }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3));

    const [requestUrl, requestInit] = fetchMock.mock.calls[0];
    expect(requestUrl).toContain('/api/v1/auth/login');
    expect(requestInit?.method).toBe('POST');
    expect(requestInit?.body).toEqual(
      JSON.stringify({ email: 'alex@example.com', password: 'sup3r-secure' })
    );

    const friendsCall = fetchMock.mock.calls.find(([callInput]) =>
      extractUrl(callInput).includes('/api/v1/friends')
    );
    expect(friendsCall?.[1]?.headers).toMatchObject({
      Authorization: `Bearer ${tokens.accessToken}`
    });

    const feedCall = fetchMock.mock.calls.find(([callInput]) =>
      extractUrl(callInput).includes('/api/v1/videos/feed')
    );
    expect(feedCall?.[1]?.headers).toMatchObject({
      Authorization: `Bearer ${tokens.accessToken}`
    });

    await screen.findByRole('heading', { name: /welcome back, alex!/i });

    const storedSession = window.localStorage.getItem('vidfriends.session');
    expect(storedSession).toBeTruthy();
    expect(storedSession).toContain('alex@example.com');
  });

  it('surfaces API errors during sign-in without clearing the form state', async () => {
    fetchMock.mockResolvedValueOnce(
      createJsonResponse(401, { error: 'Invalid credentials. Please try again.' })
    );

    const user = userEvent.setup();
    renderApp(['/login']);

    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/password/i);

    await user.type(emailInput, 'alex@example.com');
    await user.type(passwordInput, 'bad-password');
    await user.click(screen.getByRole('button', { name: /log in/i }));

    await waitFor(() => {
      expect(screen.getAllByText(/invalid credentials/i).length).toBeGreaterThan(0);
    });

    expect(emailInput).toHaveValue('alex@example.com');
    expect(passwordInput).toHaveValue('bad-password');
    expect(window.localStorage.getItem('vidfriends.session')).toBeNull();
  });

  it('creates a new account and shows the personalized dashboard welcome', async () => {
    const tokens = {
      accessToken: 'new-access-token',
      accessExpiresAt: new Date(Date.now() + 600_000).toISOString(),
      refreshToken: 'new-refresh-token',
      refreshExpiresAt: new Date(Date.now() + 86_400_000).toISOString()
    };

    fetchMock.mockImplementation((input, init) => {
      const url = extractUrl(input);
      if (url.includes('/api/v1/auth/signup')) {
        return Promise.resolve(createJsonResponse(200, { tokens }));
      }
      if (url.includes('/api/v1/friends')) {
        return Promise.resolve(createJsonResponse(200, { requests: [] }));
      }
      if (url.includes('/api/v1/videos/feed')) {
        return Promise.resolve(createJsonResponse(200, { entries: [] }));
      }
      return Promise.reject(new Error(`Unhandled request: ${url}`));
    });

    const user = userEvent.setup();
    renderApp(['/signup']);

    await user.type(screen.getByLabelText(/display name/i), 'Sam the Streamer');
    await user.type(screen.getByLabelText(/^email$/i), 'sam@example.com');
    await user.type(screen.getByLabelText(/password/i), 'strong-passphrase');
    await user.click(screen.getByRole('button', { name: /create account/i }));

    await screen.findByRole('heading', { name: /welcome back, sam the streamer!/i });
    expect(window.localStorage.getItem('vidfriends.session')).toContain('sam@example.com');
  });

  it('confirms a password reset request even when the API has not yet been implemented', async () => {
    fetchMock.mockResolvedValueOnce(createJsonResponse(404, { error: 'Not implemented' }));

    const user = userEvent.setup();
    renderApp(['/forgot-password']);

    await user.type(screen.getByLabelText(/email address/i), 'someone@example.com');
    await user.click(screen.getByRole('button', { name: /send reset link/i }));

    await screen.findByText(/you'll receive a reset link shortly/i);
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/auth/password-reset'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ email: 'someone@example.com' })
      })
    );
  });
});
