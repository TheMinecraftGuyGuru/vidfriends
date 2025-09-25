import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { Mock } from 'vitest';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { AppStateContextValue, FeedEntry } from '../../state/AppStateProvider';
import { ShareVideoForm } from '../components/ShareVideoForm';
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

describe('ShareVideoForm', () => {
  beforeEach(() => {
    useAppStateMock.mockReset();
  });

  it('prompts the user to sign in before sharing videos', () => {
    useAppStateMock.mockReturnValue(createState());

    render(<ShareVideoForm />);

    expect(screen.getByRole('button', { name: /share video/i })).toBeDisabled();
    expect(
      screen.getByText(/sign in to start sharing videos with your friends/i)
    ).toBeInTheDocument();
  });

  it('submits shares, shows progress, and resets the form on success', async () => {
    const shareVideo = vi.fn<Parameters<AppStateContextValue['shareVideo']>, Promise<FeedEntry>>();
    const mockEntry: FeedEntry = {
      id: 'entry-1',
      title: 'Incredible speedruns',
      sharedBy: 'Alex',
      sharedAt: new Date().toISOString(),
      platform: 'YouTube',
      url: 'https://example.com/watch?v=abc',
      thumbnailUrl: 'https://example.com/thumb.jpg',
      channelName: 'Speed Central',
      durationSeconds: 512,
      viewCount: 100_000,
      description: 'The best runs of the week.',
      tags: ['games'],
      reactions: { like: 0, love: 0, wow: 0, laugh: 0 },
      userReaction: null
    };

    let resolveShare!: (entry: FeedEntry) => void;
    shareVideo.mockReturnValue(
      new Promise<FeedEntry>((resolve) => {
        resolveShare = resolve;
      })
    );

    useAppStateMock.mockReturnValue(
      createState({
        auth: {
          user: { id: 'user-1', email: 'user@example.com', displayName: 'User One' },
          status: 'authenticated',
          tokens: null
        },
        shareVideo
      })
    );

    const user = userEvent.setup();
    render(<ShareVideoForm />);

    const urlInput = screen.getByLabelText(/video url/i);
    await user.type(urlInput, 'https://example.com/watch?v=abc');
    await user.type(screen.getByLabelText(/optional title override/i), 'My run');

    const submitButton = screen.getByRole('button', { name: /share video/i });
    expect(submitButton).toBeEnabled();

    await user.click(submitButton);

    const pendingButton = await screen.findByRole('button', { name: /sharing/i });
    expect(pendingButton).toBeDisabled();
    expect(await screen.findByText(/fetching video details/i)).toBeInTheDocument();

    resolveShare(mockEntry);

    await waitFor(() =>
      expect(shareVideo).toHaveBeenCalledWith({ title: 'My run', url: 'https://example.com/watch?v=abc' })
    );
    await screen.findByText(/shared “incredible speedruns” with your friends/i);

    expect(urlInput).toHaveValue('');
    expect(screen.getByLabelText(/optional title override/i)).toHaveValue('');
    expect(screen.getByRole('button', { name: /share video/i })).toBeDisabled();
  });

  it('surfaces backend errors and keeps the form state intact', async () => {
    const shareVideo = vi
      .fn<Parameters<AppStateContextValue['shareVideo']>, Promise<FeedEntry>>()
      .mockRejectedValue(new Error('Unable to reach the server.'));

    useAppStateMock.mockReturnValue(
      createState({
        auth: {
          user: { id: 'user-1', email: 'user@example.com', displayName: 'User One' },
          status: 'authenticated',
          tokens: null
        },
        shareVideo
      })
    );

    const user = userEvent.setup();
    render(<ShareVideoForm />);

    const urlInput = screen.getByLabelText(/video url/i);
    await user.type(urlInput, 'https://example.com/watch?v=abc');

    await user.click(screen.getByRole('button', { name: /share video/i }));

    await waitFor(() =>
      expect(screen.getByText(/unable to reach the server/i)).toBeInTheDocument()
    );

    expect(shareVideo).toHaveBeenCalled();
    expect(urlInput).toHaveValue('https://example.com/watch?v=abc');
    expect(screen.getByRole('button', { name: /share video/i })).toBeEnabled();
  });
});
