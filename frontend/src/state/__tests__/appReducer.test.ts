import { describe, expect, it } from 'vitest';

import type { AppStateAction, FeedEntry, FriendConnection, FriendInvite } from '../AppStateProvider';
import { appReducer } from '../AppStateProvider';

type AppState = Parameters<typeof appReducer>[0];

function createState(overrides: Partial<AppState> = {}): AppState {
  return {
    auth: { user: null, status: 'anonymous', tokens: null },
    friends: { pending: [], connections: [] },
    feed: { entries: [] },
    ...overrides
  };
}

describe('appReducer', () => {
  it('promotes accepted invites into the friend list and removes declined ones', () => {
    const invite: FriendInvite = { id: 'inv-1', displayName: 'Casey', mutualFriends: 3 };
    const existingFriend: FriendConnection = {
      id: 'friend-1',
      displayName: 'Rowan',
      status: 'online'
    };

    const state = createState({
      friends: {
        pending: [invite],
        connections: [existingFriend]
      }
    });

    const accepted = appReducer(state, {
      type: 'resolve-invite',
      payload: { inviteId: 'inv-1', accepted: true }
    } satisfies AppStateAction);

    expect(accepted.friends.pending).toHaveLength(0);
    expect(accepted.friends.connections[0]).toMatchObject({
      id: invite.id,
      displayName: invite.displayName,
      status: 'offline'
    });

    const declined = appReducer(state, {
      type: 'resolve-invite',
      payload: { inviteId: 'inv-1', accepted: false }
    } satisfies AppStateAction);

    expect(declined.friends.pending).toHaveLength(0);
    expect(declined.friends.connections).toHaveLength(1);
    expect(declined.friends.connections[0]).toEqual(existingFriend);
  });

  it('adds optimistic shares to the beginning of the feed and replaces them on success', () => {
    const originalEntry: FeedEntry = {
      id: 'feed-1',
      title: 'Cozy builds',
      sharedBy: 'Alex',
      sharedAt: new Date(Date.now() - 60_000).toISOString(),
      platform: 'YouTube',
      url: 'https://example.com/1',
      thumbnailUrl: 'https://example.com/thumb1.jpg',
      channelName: 'Builders Guild',
      durationSeconds: 480,
      viewCount: 12_300,
      description: 'A calming timelapse build.',
      tags: ['Relax'],
      reactions: { like: 2, love: 0, wow: 0, laugh: 0 },
      userReaction: null
    };
    const optimisticEntry: FeedEntry = {
      ...originalEntry,
      id: 'optimistic-1',
      title: 'Speedrun tips',
      sharedAt: new Date().toISOString()
    };
    const persistedEntry: FeedEntry = {
      ...optimisticEntry,
      id: 'feed-2',
      sharedAt: new Date().toISOString(),
      thumbnailUrl: 'https://example.com/thumb2.jpg'
    };

    const state = createState({ feed: { entries: [originalEntry] } });

    const withOptimistic = appReducer(state, {
      type: 'share-video-start',
      payload: optimisticEntry
    } satisfies AppStateAction);

    expect(withOptimistic.feed.entries.map((entry) => entry.id)).toEqual([
      'optimistic-1',
      'feed-1'
    ]);

    const withPersisted = appReducer(withOptimistic, {
      type: 'share-video-success',
      payload: { optimisticId: 'optimistic-1', entry: persistedEntry }
    } satisfies AppStateAction);

    expect(withPersisted.feed.entries.map((entry) => entry.id)).toEqual(['feed-2', 'feed-1']);
    expect(withPersisted.feed.entries[0]).toEqual(persistedEntry);
  });

  it('removes optimistic shares when persistence fails', () => {
    const optimisticEntry: FeedEntry = {
      id: 'optimistic-1',
      title: 'Speedrun tips',
      sharedBy: 'Alex',
      sharedAt: new Date().toISOString(),
      platform: 'YouTube',
      url: 'https://example.com/2',
      thumbnailUrl: 'https://example.com/thumb2.jpg',
      channelName: 'Builders Guild',
      durationSeconds: 480,
      viewCount: 12_300,
      description: 'A calming timelapse build.',
      tags: ['Relax'],
      reactions: { like: 2, love: 0, wow: 0, laugh: 0 },
      userReaction: null
    };

    const state = createState({ feed: { entries: [] } });

    const withOptimistic = appReducer(state, {
      type: 'share-video-start',
      payload: optimisticEntry
    } satisfies AppStateAction);

    expect(withOptimistic.feed.entries).toHaveLength(1);

    const reverted = appReducer(withOptimistic, {
      type: 'share-video-error',
      payload: { optimisticId: 'optimistic-1' }
    } satisfies AppStateAction);

    expect(reverted.feed.entries).toHaveLength(0);
  });

  it('toggles reactions and keeps counts consistent', () => {
    const entry: FeedEntry = {
      id: 'feed-1',
      title: 'Morning yoga',
      sharedBy: 'Jamie',
      sharedAt: new Date().toISOString(),
      platform: 'YouTube',
      url: 'https://example.com/yoga',
      thumbnailUrl: 'https://example.com/thumb-yoga.jpg',
      channelName: 'Peaceful Moves',
      durationSeconds: 1_800,
      viewCount: 91_500,
      description: 'Start your day with a gentle routine.',
      tags: ['Wellness'],
      reactions: { like: 10, love: 5, wow: 1, laugh: 0 },
      userReaction: 'love'
    };

    const state = createState({ feed: { entries: [entry] } });

    const toggledOff = appReducer(state, {
      type: 'react-to-video',
      payload: { entryId: 'feed-1', reaction: 'love' }
    } satisfies AppStateAction);

    expect(toggledOff.feed.entries[0].reactions.love).toBe(4);
    expect(toggledOff.feed.entries[0].userReaction).toBeNull();

    const switched = appReducer(toggledOff, {
      type: 'react-to-video',
      payload: { entryId: 'feed-1', reaction: 'wow' }
    } satisfies AppStateAction);

    expect(switched.feed.entries[0].reactions).toMatchObject({
      like: 10,
      love: 4,
      wow: 2,
      laugh: 0
    });
    expect(switched.feed.entries[0].userReaction).toBe('wow');
  });
});
