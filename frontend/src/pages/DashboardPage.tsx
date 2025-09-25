import { useMemo } from 'react';

import { useAppState } from '../state/useAppState';
import { FriendListManager } from './components/FriendListManager';
import { ShareVideoForm } from './components/ShareVideoForm';
import { VideoFeedPanel } from './components/VideoFeedPanel';

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

      <ShareVideoForm />

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'minmax(0, 360px) minmax(0, 1fr)',
          gap: '2rem',
          alignItems: 'start'
        }}
      >
        <FriendListManager />

        <VideoFeedPanel feed={feed} />
      </div>
    </section>
  );
}
