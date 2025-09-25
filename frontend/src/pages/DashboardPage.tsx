import { useMemo } from 'react';

import { useAppState } from '../state/useAppState';

export function DashboardPage() {
  const { auth, friends, feed } = useAppState();

  const welcomeTitle = useMemo(() => {
    if (auth.user) {
      return `Welcome back, ${auth.user.displayName}!`;
    }
    return 'Welcome to your VidFriends dashboard';
  }, [auth.user]);

  return (
    <section>
      <header>
        <h2>{welcomeTitle}</h2>
        <p>
          Track friend activity, respond to invitations, and explore what your
          circle is watching.
        </p>
      </header>

      <article style={{ marginTop: '2rem' }}>
        <h3>Friend invitations</h3>
        {friends.pending.length === 0 ? (
          <p>No pending invites.</p>
        ) : (
          <ul>
            {friends.pending.map((invite) => (
              <li key={invite.id}>{invite.displayName}</li>
            ))}
          </ul>
        )}
      </article>

      <article style={{ marginTop: '2rem' }}>
        <h3>Shared videos</h3>
        {feed.entries.length === 0 ? (
          <p>Share a link to see it appear here.</p>
        ) : (
          <ul style={{ listStyle: 'none', padding: 0, display: 'grid', gap: '1rem' }}>
            {feed.entries.map((entry) => (
              <li
                key={entry.id}
                style={{
                  backgroundColor: 'rgba(15, 23, 42, 0.85)',
                  padding: '1rem',
                  borderRadius: '0.75rem'
                }}
              >
                <p style={{ fontWeight: 600 }}>{entry.title}</p>
                <p style={{ fontSize: '0.85rem', color: 'rgba(148, 163, 184, 0.9)' }}>
                  Shared by {entry.sharedBy}
                </p>
              </li>
            ))}
          </ul>
        )}
      </article>
    </section>
  );
}
