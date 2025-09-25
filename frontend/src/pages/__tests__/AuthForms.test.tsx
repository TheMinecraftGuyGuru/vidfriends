import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import type { Mock } from 'vitest';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { AppStateContextValue } from '../../state/AppStateProvider';
import { ForgotPasswordPage } from '../ForgotPasswordPage';
import { LoginPage } from '../LoginPage';
import { SignupPage } from '../SignupPage';
import { useAppState } from '../../state/useAppState';

vi.mock('../../state/useAppState', () => ({
  useAppState: vi.fn()
}));

const useAppStateMock = useAppState as unknown as Mock;

function createState(overrides: Partial<AppStateContextValue> = {}): AppStateContextValue {
  return {
    auth: { user: null, status: 'anonymous', tokens: null },
    friends: { pending: [], connections: [] },
    feed: { entries: [] },
    signIn: vi.fn(),
    signUp: vi.fn(),
    requestPasswordReset: vi.fn(),
    signOut: vi.fn(),
    acceptInvite: vi.fn(),
    declineInvite: vi.fn(),
    respondToInvite: vi.fn(),
    shareVideo: vi.fn(),
    reactToVideo: vi.fn(),
    ...overrides
  };
}

describe('Authentication form components', () => {
  beforeEach(() => {
    useAppStateMock.mockReset();
  });

  it('invokes the sign-in flow and shows errors on failure', async () => {
    const signIn = vi
      .fn<AppStateContextValue['signIn']>()
      .mockRejectedValue(new Error('Invalid credentials.'));

    useAppStateMock.mockReturnValue(createState({ signIn }));

    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    await user.type(screen.getByLabelText(/email/i), 'alex@example.com');
    await user.type(screen.getByLabelText(/password/i), 'bad-pass');
    await user.click(screen.getByRole('button', { name: /log in/i }));

    await waitFor(() => expect(signIn).toHaveBeenCalledWith({ email: 'alex@example.com', password: 'bad-pass' }));
    await screen.findByText(/invalid credentials/i);
    expect(screen.getByRole('button', { name: /log in/i })).toBeEnabled();
  });

  it('submits the sign-up form with the provided details', async () => {
    const signUp = vi.fn<AppStateContextValue['signUp']>().mockResolvedValue({
      id: 'alex@example.com',
      email: 'alex@example.com',
      displayName: 'Alex'
    });

    useAppStateMock.mockReturnValue(createState({ signUp }));

    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <SignupPage />
      </MemoryRouter>
    );

    await user.type(screen.getByLabelText(/display name/i), 'Alex');
    await user.type(screen.getByLabelText(/^email$/i), 'alex@example.com');
    await user.type(screen.getByLabelText(/password/i), 'pa55word');

    await user.click(screen.getByRole('button', { name: /create account/i }));

    await waitFor(() =>
      expect(signUp).toHaveBeenCalledWith({
        displayName: 'Alex',
        email: 'alex@example.com',
        password: 'pa55word'
      })
    );

    expect(screen.getByRole('button', { name: /create account/i })).toBeEnabled();
  });

  it('requests a password reset using the entered email', async () => {
    const requestPasswordReset = vi
      .fn<AppStateContextValue['requestPasswordReset']>()
      .mockResolvedValue();

    useAppStateMock.mockReturnValue(createState({ requestPasswordReset }));

    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <ForgotPasswordPage />
      </MemoryRouter>
    );

    await user.type(screen.getByLabelText(/email address/i), 'friend@example.com');
    await user.click(screen.getByRole('button', { name: /send reset link/i }));

    await waitFor(() => expect(requestPasswordReset).toHaveBeenCalledWith('friend@example.com'));
    await screen.findByText(/you'll receive a reset link shortly/i);
  });
});
