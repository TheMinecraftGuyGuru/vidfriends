import { useMemo } from 'react';

import { useAppState } from '../state/useAppState';
import { FriendListManager } from './components/FriendListManager';

export function DashboardPage() {
  const { auth, feed } = useAppState();

  const welcomeTitle = useMemo(() => {
    if (auth.user) {
      return `Welcome back, ${auth.user.displayName}!`;
    }
    return 'Welcome to your VidFriends dashboard';
  }, [auth.user]);

  return (
    <section style={{ display: 'flex', flexDirection: 'column', gap: '2.5rem' }}>
      <header>
        <h2 style={{ marginBottom: '0.5rem' }}>{welcomeTitle}</h2>
        <p style={{ color: 'rgba(148, 163, 184, 0.9)', maxWidth: '60ch' }}>
          Track friend activity, respond to invitations, and explore what your circle is watching.
        </p>
      </header>

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'minmax(0, 360px) minmax(0, 1fr)',
          gap: '2rem',
          alignItems: 'start'
        }}
      >
        <FriendListManager />

        <article
          aria-label="Shared video feed"
          style={{
            backgroundColor: 'rgba(15, 23, 42, 0.85)',
            padding: '1.75rem',
            borderRadius: '1rem',
            display: 'flex',
            flexDirection: 'column',
            gap: '1.25rem',
            boxShadow: '0 25px 40px -24px rgba(15, 23, 42, 0.7)'
          }}
        >
          <div>
            <h3 style={{ marginBottom: '0.5rem' }}>Shared videos</h3>
            <p style={{ color: 'rgba(148, 163, 184, 0.85)', fontSize: '0.9rem' }}>
              Keep up with the latest videos from friends and jump into conversations.
            </p>
          </div>
          {feed.entries.length === 0 ? (
            <p style={{ color: 'rgba(148, 163, 184, 0.85)' }}>
              Share a link to see it appear here.
            </p>
          ) : (
            <ul style={{ listStyle: 'none', padding: 0, display: 'grid', gap: '1rem' }}>
              {feed.entries.map((entry) => (
                <li
                  key={entry.id}
                  style={{
                    backgroundColor: 'rgba(15, 23, 42, 0.6)',
                    border: '1px solid rgba(148, 163, 184, 0.2)',
                    borderRadius: '0.85rem',
                    padding: '1rem 1.25rem',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '0.35rem'
                  }}
                >
                  <p style={{ fontWeight: 600 }}>{entry.title}</p>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                    <p style={{ fontSize: '0.85rem', color: 'rgba(148, 163, 184, 0.9)' }}>
                      Shared by {entry.sharedBy}
                    </p>
                    <time
                      dateTime={entry.sharedAt}
                      style={{ fontSize: '0.75rem', color: 'rgba(148, 163, 184, 0.7)' }}
                    >
                      {new Date(entry.sharedAt).toLocaleString()}
                    </time>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </article>
      </div>
    </section>
  );
}
