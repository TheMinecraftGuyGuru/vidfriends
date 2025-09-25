import { useMemo, useState } from 'react';

import type { FeedEntry, ReactionType } from '../../state/AppStateProvider';
import { useAppState } from '../../state/useAppState';

type FeedFilter = 'all' | 'recent' | 'friends-online' | 'trending' | `tag:${string}`;

type VideoFeedPanelProps = {
  feed: { entries: FeedEntry[] };
};

const reactionConfig: Array<{ type: ReactionType; label: string; icon: string }> = [
  { type: 'like', label: 'Like', icon: '👍' },
  { type: 'love', label: 'Love', icon: '❤️' },
  { type: 'wow', label: 'Wow', icon: '🤯' },
  { type: 'laugh', label: 'Laugh', icon: '😂' }
];

export function VideoFeedPanel({ feed }: VideoFeedPanelProps) {
  const { friends, reactToVideo } = useAppState();
  const [activeFilter, setActiveFilter] = useState<FeedFilter>('all');

  const onlineFriendNames = useMemo(
    () => new Set(friends.connections.filter((friend) => friend.status !== 'offline').map((friend) => friend.displayName)),
    [friends.connections]
  );

  const filterOptions = useMemo(() => {
    const uniqueTags = new Set<string>();
    feed.entries.forEach((entry) => {
      entry.tags.forEach((tag) => uniqueTags.add(tag));
    });
    const tags = Array.from(uniqueTags).sort((a, b) => a.localeCompare(b));
    return [
      { id: 'all' as FeedFilter, label: 'All' },
      { id: 'recent' as FeedFilter, label: 'Recent' },
      { id: 'trending' as FeedFilter, label: 'Trending' },
      { id: 'friends-online' as FeedFilter, label: 'From Online Friends' },
      ...tags.map((tag) => ({ id: `tag:${tag}` as FeedFilter, label: tag }))
    ];
  }, [feed.entries]);

  const filteredEntries = useMemo(() => {
    const entries = [...feed.entries];
    const now = Date.now();

    switch (activeFilter) {
      case 'recent':
        return entries
          .filter((entry) => now - new Date(entry.sharedAt).getTime() <= 1000 * 60 * 60 * 24 * 2)
          .sort((a, b) => new Date(b.sharedAt).getTime() - new Date(a.sharedAt).getTime());
      case 'trending':
        return entries
          .filter((entry) => totalReactions(entry) >= 10)
          .sort((a, b) => totalReactions(b) - totalReactions(a));
      case 'friends-online':
        return entries
          .filter((entry) => onlineFriendNames.has(entry.sharedBy))
          .sort((a, b) => new Date(b.sharedAt).getTime() - new Date(a.sharedAt).getTime());
      default:
        if (activeFilter.startsWith('tag:')) {
          const tag = activeFilter.replace('tag:', '');
          return entries
            .filter((entry) => entry.tags.includes(tag))
            .sort((a, b) => new Date(b.sharedAt).getTime() - new Date(a.sharedAt).getTime());
        }
        return entries.sort((a, b) => new Date(b.sharedAt).getTime() - new Date(a.sharedAt).getTime());
    }
  }, [activeFilter, feed.entries, onlineFriendNames]);

  return (
    <article
      aria-label="Shared video feed"
      style={{
        backgroundColor: 'rgba(15, 23, 42, 0.85)',
        padding: '1.75rem',
        borderRadius: '1rem',
        display: 'flex',
        flexDirection: 'column',
        gap: '1.75rem',
        boxShadow: '0 25px 40px -24px rgba(15, 23, 42, 0.7)'
      }}
    >
      <header style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        <div>
          <h3 style={{ marginBottom: '0.5rem' }}>Shared videos</h3>
          <p style={{ color: 'rgba(148, 163, 184, 0.85)', fontSize: '0.9rem', maxWidth: '65ch' }}>
            Dive into what your friends are sharing. Browse by momentum, who is online now, or the topics you care about most.
          </p>
        </div>

        <nav aria-label="Filter shared videos" style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
          {filterOptions.map((filter) => (
            <button
              key={filter.id}
              type="button"
              onClick={() => setActiveFilter(filter.id)}
              style={{
                borderRadius: '9999px',
                border: '1px solid rgba(148, 163, 184, 0.25)',
                padding: '0.4rem 0.85rem',
                backgroundColor: filter.id === activeFilter ? 'rgba(96, 165, 250, 0.16)' : 'rgba(15, 23, 42, 0.6)',
                color: filter.id === activeFilter ? '#f8fafc' : 'rgba(148, 163, 184, 0.9)',
                fontSize: '0.85rem',
                transition: 'background-color 120ms ease, color 120ms ease'
              }}
            >
              {filter.label}
            </button>
          ))}
        </nav>
      </header>

      {filteredEntries.length === 0 ? (
        <p style={{ color: 'rgba(148, 163, 184, 0.85)' }}>
          Nothing here yet. Share a link or try a different filter to explore what your crew is watching.
        </p>
      ) : (
        <ul style={{ listStyle: 'none', padding: 0, display: 'grid', gap: '1.25rem' }}>
          {filteredEntries.map((entry) => (
            <li key={entry.id}>
              <VideoFeedCard entry={entry} onReact={reactToVideo} />
            </li>
          ))}
        </ul>
      )}
    </article>
  );
}

type VideoFeedCardProps = {
  entry: FeedEntry;
  onReact: (entryId: string, reaction: ReactionType) => void;
};

function VideoFeedCard({ entry, onReact }: VideoFeedCardProps) {
  const total = totalReactions(entry);
  const formattedDuration = formatDuration(entry.durationSeconds);
  const formattedViews = new Intl.NumberFormat('en', { notation: 'compact' }).format(entry.viewCount);

  return (
    <article
      style={{
        backgroundColor: 'rgba(15, 23, 42, 0.6)',
        border: '1px solid rgba(148, 163, 184, 0.2)',
        borderRadius: '1rem',
        padding: '1.25rem',
        display: 'grid',
        gap: '1rem',
        gridTemplateColumns: 'minmax(0, 240px) minmax(0, 1fr)'
      }}
    >
      <a
        href={entry.url}
        target="_blank"
        rel="noreferrer"
        style={{
          position: 'relative',
          borderRadius: '0.85rem',
          overflow: 'hidden',
          display: 'block',
          backgroundColor: '#0f172a'
        }}
      >
        <img
          src={entry.thumbnailUrl}
          alt={`Preview for ${entry.title}`}
          style={{ width: '100%', height: '100%', objectFit: 'cover' }}
          loading="lazy"
        />
        <span
          style={{
            position: 'absolute',
            top: '0.75rem',
            left: '0.75rem',
            backgroundColor: 'rgba(15, 23, 42, 0.75)',
            color: '#e2e8f0',
            padding: '0.2rem 0.6rem',
            borderRadius: '9999px',
            fontSize: '0.7rem',
            fontWeight: 600
          }}
        >
          {entry.platform}
        </span>
        <span
          style={{
            position: 'absolute',
            bottom: '0.75rem',
            right: '0.75rem',
            backgroundColor: 'rgba(15, 23, 42, 0.75)',
            color: '#f8fafc',
            padding: '0.2rem 0.6rem',
            borderRadius: '0.35rem',
            fontSize: '0.75rem',
            fontVariantNumeric: 'tabular-nums'
          }}
        >
          {formattedDuration}
        </span>
      </a>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.65rem' }}>
        <header style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', gap: '1rem', flexWrap: 'wrap' }}>
            <h4 style={{ margin: 0 }}>{entry.title}</h4>
            <time
              dateTime={entry.sharedAt}
              style={{ fontSize: '0.75rem', color: 'rgba(148, 163, 184, 0.7)' }}
              title={new Date(entry.sharedAt).toLocaleString()}
            >
              {formatRelativeTime(entry.sharedAt)}
            </time>
          </div>
          <p style={{ fontSize: '0.85rem', color: 'rgba(148, 163, 184, 0.85)' }}>
            Shared by <strong style={{ color: '#f8fafc' }}>{entry.sharedBy}</strong> · {entry.channelName}
          </p>
        </header>

        <p style={{ color: 'rgba(203, 213, 225, 0.9)', fontSize: '0.9rem', lineHeight: 1.5 }}>
          {entry.description}
        </p>

        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
          {entry.tags.map((tag) => (
            <span
              key={tag}
              style={{
                fontSize: '0.7rem',
                padding: '0.25rem 0.65rem',
                backgroundColor: 'rgba(96, 165, 250, 0.1)',
                border: '1px solid rgba(96, 165, 250, 0.15)',
                borderRadius: '999px',
                color: 'rgba(191, 219, 254, 0.95)'
              }}
            >
              #{tag}
            </span>
          ))}
        </div>

        <footer
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: '0.75rem'
          }}
        >
          <div style={{ display: 'flex', gap: '0.35rem', flexWrap: 'wrap' }}>
            {reactionConfig.map((reaction) => {
              const isActive = entry.userReaction === reaction.type;
              return (
                <button
                  key={reaction.type}
                  type="button"
                  onClick={() => onReact(entry.id, reaction.type)}
                  aria-pressed={isActive}
                  title={`${reaction.label} reaction`}
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '0.35rem',
                    borderRadius: '9999px',
                    border: '1px solid rgba(148, 163, 184, 0.25)',
                    backgroundColor: isActive ? 'rgba(59, 130, 246, 0.2)' : 'rgba(15, 23, 42, 0.5)',
                    color: isActive ? '#f8fafc' : 'rgba(203, 213, 225, 0.95)',
                    padding: '0.35rem 0.75rem',
                    fontSize: '0.8rem',
                    cursor: 'pointer',
                    transition: 'background-color 120ms ease, transform 120ms ease'
                  }}
                >
                  <span aria-hidden="true">{reaction.icon}</span>
                  <span>{entry.reactions[reaction.type] ?? 0}</span>
                </button>
              );
            })}
          </div>

          <p style={{ margin: 0, fontSize: '0.8rem', color: 'rgba(148, 163, 184, 0.85)' }}>
            {formattedViews} views · {total} reactions
          </p>
        </footer>
      </div>
    </article>
  );
}

function totalReactions(entry: FeedEntry) {
  return Object.values(entry.reactions).reduce((sum, value) => sum + (value ?? 0), 0);
}

function formatDuration(durationSeconds: number) {
  const minutes = Math.floor(durationSeconds / 60);
  const seconds = durationSeconds % 60;
  const hours = Math.floor(minutes / 60);
  const minutesWithinHour = minutes % 60;

  const parts = [] as string[];
  if (hours > 0) {
    parts.push(String(hours));
    parts.push(minutesWithinHour.toString().padStart(2, '0'));
  } else {
    parts.push(minutes.toString());
  }
  parts.push(seconds.toString().padStart(2, '0'));
  return parts.join(':');
}

function formatRelativeTime(timestamp: string) {
  const diff = Date.now() - new Date(timestamp).getTime();
  const units: Array<[Intl.RelativeTimeFormatUnit, number]> = [
    ['day', 1000 * 60 * 60 * 24],
    ['hour', 1000 * 60 * 60],
    ['minute', 1000 * 60],
    ['second', 1000]
  ];
  const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

  for (const [unit, ms] of units) {
    if (Math.abs(diff) >= ms || unit === 'second') {
      const value = Math.round(diff / ms);
      return rtf.format(-value, unit);
    }
  }
  return 'just now';
}
