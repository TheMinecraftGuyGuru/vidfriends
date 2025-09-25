import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { Mock } from 'vitest';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { AppStateContextValue, FeedEntry } from '../../state/AppStateProvider';
import { VideoFeedPanel } from '../components/VideoFeedPanel';
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

const baseEntries: FeedEntry[] = [
  {
    id: 'feed-1',
    title: 'Top 10 Cozy Indie Games',
    sharedBy: 'Rowan Carter',
    sharedAt: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
    platform: 'YouTube',
    url: 'https://example.com/indie',
    thumbnailUrl: 'https://example.com/thumb1.jpg',
    channelName: 'Indie Realm',
    durationSeconds: 900,
    viewCount: 42_000,
    description: 'Discover new cozy titles perfect for winding down after a long day.',
    tags: ['Games', 'Cozy'],
    reactions: { like: 18, love: 6, wow: 2, laugh: 1 },
    userReaction: null
  },
  {
    id: 'feed-2',
    title: 'How the JWST Sees the Universe',
    sharedBy: 'Miguel Chen',
    sharedAt: new Date(Date.now() - 1000 * 60 * 60 * 12).toISOString(),
    platform: 'Nebula',
    url: 'https://example.com/jwst',
    thumbnailUrl: 'https://example.com/thumb2.jpg',
    channelName: 'Cosmic Perspectives',
    durationSeconds: 1245,
    viewCount: 38_700,
    description: 'Astrophysicist Dr. Mei Park breaks down the science behind the telescope.',
    tags: ['Science', 'Space'],
    reactions: { like: 21, love: 11, wow: 7, laugh: 0 },
    userReaction: 'love'
  }
];

describe('VideoFeedPanel', () => {
  beforeEach(() => {
    useAppStateMock.mockReset();
  });

  it('filters entries using the provided controls', async () => {
    const user = userEvent.setup();
    useAppStateMock.mockReturnValue(
      createState({
        friends: {
          pending: [],
          connections: [
            { id: 'friend-1', displayName: 'Rowan Carter', status: 'online' },
            { id: 'friend-2', displayName: 'Miguel Chen', status: 'offline' }
          ]
        }
      })
    );

    render(<VideoFeedPanel feed={{ entries: baseEntries }} />);

    expect(screen.getByText(/top 10 cozy indie games/i)).toBeInTheDocument();
    expect(screen.getByText(/how the jwst sees the universe/i)).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /from online friends/i }));

    expect(screen.getByText(/top 10 cozy indie games/i)).toBeInTheDocument();
    expect(screen.queryByText(/how the jwst sees the universe/i)).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /science/i }));

    expect(screen.getByText(/how the jwst sees the universe/i)).toBeInTheDocument();
    expect(screen.queryByText(/top 10 cozy indie games/i)).not.toBeInTheDocument();
  });

  it('notifies the parent when a reaction is clicked', async () => {
    const user = userEvent.setup();
    const reactToVideo = vi.fn();
    useAppStateMock.mockReturnValue(
      createState({
        friends: { pending: [], connections: [] },
        reactToVideo
      })
    );

    render(<VideoFeedPanel feed={{ entries: baseEntries }} />);

    const [reactionButton] = screen.getAllByTitle('Like reaction');

    await user.click(reactionButton);
    expect(reactToVideo).toHaveBeenCalledWith('feed-1', 'like');

    await user.click(reactionButton);
    expect(reactToVideo).toHaveBeenLastCalledWith('feed-1', 'like');
  });

  it('shows an empty state when there are no entries to display', () => {
    useAppStateMock.mockReturnValue(createState());

    render(<VideoFeedPanel feed={{ entries: [] }} />);

    expect(
      screen.getByText(/nothing here yet. share a link or try a different filter/i)
    ).toBeInTheDocument();
  });
});
