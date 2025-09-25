import { useEffect, useMemo, useState } from 'react';

import type { FriendConnection, FriendInvite } from '../../state/AppStateProvider';
import { useAppState } from '../../state/useAppState';

const statusColors: Record<FriendConnection['status'], string> = {
  online: '#10b981',
  away: '#f59e0b',
  offline: '#64748b'
};

const statusLabels: Record<FriendConnection['status'], string> = {
  online: 'Online',
  away: 'Away',
  offline: 'Offline'
};

interface FriendListManagerViewProps {
  friends: { pending: FriendInvite[]; connections: FriendConnection[] };
  respondToInvite: (inviteId: string, accepted: boolean) => Promise<void>;
}

export function FriendListManagerView({ friends, respondToInvite }: FriendListManagerViewProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [feedbackMessage, setFeedbackMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [inviteActions, setInviteActions] = useState<Record<string, 'accept' | 'decline'>>({});

  useEffect(() => {
    if (!feedbackMessage) {
      return undefined;
    }
    const timeout = setTimeout(() => setFeedbackMessage(null), 4000);
    return () => clearTimeout(timeout);
  }, [feedbackMessage]);

  useEffect(() => {
    if (!errorMessage) {
      return undefined;
    }
    const timeout = setTimeout(() => setErrorMessage(null), 6000);
    return () => clearTimeout(timeout);
  }, [errorMessage]);

  const filteredConnections = useMemo(() => {
    if (!searchTerm.trim()) {
      return friends.connections;
    }
    const normalized = searchTerm.trim().toLowerCase();
    return friends.connections.filter((connection) =>
      connection.displayName.toLowerCase().includes(normalized)
    );
  }, [friends.connections, searchTerm]);

  async function handleInviteResponse(inviteId: string, accepted: boolean) {
    setFeedbackMessage(null);
    setErrorMessage(null);
    setInviteActions((prev) => ({
      ...prev,
      [inviteId]: accepted ? 'accept' : 'decline'
    }));
    try {
      await respondToInvite(inviteId, accepted);
      setFeedbackMessage(accepted ? 'Invitation accepted.' : 'Invitation declined.');
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Something went wrong. Please retry.');
    } finally {
      setInviteActions((prev) => {
        const next = { ...prev };
        delete next[inviteId];
        return next;
      });
    }
  }

  const hasPendingInvites = friends.pending.length > 0;
  const inviteDescription = hasPendingInvites
    ? 'Respond to new friend invitations.'
    : "You're all caught up on friend requests.";

  return (
    <section
      aria-label="Manage friends"
      style={{
        backgroundColor: 'rgba(15, 23, 42, 0.85)',
        padding: '1.75rem',
        borderRadius: '1rem',
        display: 'flex',
        flexDirection: 'column',
        gap: '1.5rem',
        boxShadow: '0 25px 40px -24px rgba(15, 23, 42, 0.7)'
      }}
    >
      <header>
        <h3 style={{ marginBottom: '0.5rem' }}>Friends</h3>
        <p style={{ color: 'rgba(148, 163, 184, 0.9)', fontSize: '0.9rem' }}>{inviteDescription}</p>
      </header>

      <div aria-live="polite" style={{ minHeight: '1.2rem' }}>
        {feedbackMessage && (
          <p style={{ color: '#22c55e', fontSize: '0.9rem' }}>{feedbackMessage}</p>
        )}
        {errorMessage && (
          <p role="alert" style={{ color: '#f87171', fontSize: '0.9rem' }}>
            {errorMessage}
          </p>
        )}
      </div>

      <section aria-label="Pending invitations" style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
          <h4 style={{ margin: 0 }}>Pending invitations</h4>
          <span style={{ fontSize: '0.8rem', color: 'rgba(148, 163, 184, 0.8)' }}>
            {friends.pending.length} pending
          </span>
        </div>
        {hasPendingInvites ? (
          <ul style={{ listStyle: 'none', padding: 0, display: 'grid', gap: '0.75rem' }}>
            {friends.pending.map((invite) => {
              const actionState = inviteActions[invite.id];
              const accepting = actionState === 'accept';
              const declining = actionState === 'decline';
              return (
                <li
                  key={invite.id}
                  style={{
                    backgroundColor: 'rgba(15, 23, 42, 0.6)',
                    border: '1px solid rgba(148, 163, 184, 0.2)',
                    borderRadius: '0.9rem',
                    padding: '1rem 1.25rem',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    gap: '1rem'
                  }}
                >
                  <div>
                    <p style={{ fontWeight: 600, marginBottom: '0.25rem' }}>{invite.displayName}</p>
                    <p style={{ color: 'rgba(148, 163, 184, 0.8)', fontSize: '0.85rem' }}>
                      {invite.mutualFriends} mutual friend{invite.mutualFriends === 1 ? '' : 's'}
                    </p>
                  </div>
                  <div style={{ display: 'flex', gap: '0.75rem' }}>
                    <button
                      type="button"
                      onClick={() => handleInviteResponse(invite.id, true)}
                      disabled={accepting || declining}
                      aria-label={`Accept ${invite.displayName}`}
                      style={{
                        backgroundColor: '#22c55e',
                        color: '#0f172a',
                        fontWeight: 600,
                        borderRadius: '9999px',
                        border: 'none',
                        padding: '0.5rem 1.25rem',
                        cursor: accepting || declining ? 'not-allowed' : 'pointer',
                        opacity: accepting || declining ? 0.65 : 1
                      }}
                    >
                      {accepting ? 'Accepting…' : 'Accept'}
                    </button>
                    <button
                      type="button"
                      onClick={() => handleInviteResponse(invite.id, false)}
                      disabled={accepting || declining}
                      aria-label={`Decline ${invite.displayName}`}
                      style={{
                        backgroundColor: 'transparent',
                        color: '#f87171',
                        fontWeight: 600,
                        borderRadius: '9999px',
                        border: '1px solid rgba(248, 113, 113, 0.65)',
                        padding: '0.5rem 1.25rem',
                        cursor: accepting || declining ? 'not-allowed' : 'pointer',
                        opacity: accepting || declining ? 0.6 : 1
                      }}
                    >
                      {declining ? 'Declining…' : 'Decline'}
                    </button>
                  </div>
                </li>
              );
            })}
          </ul>
        ) : (
          <p style={{ color: 'rgba(148, 163, 184, 0.8)', fontSize: '0.9rem' }}>No pending invites.</p>
        )}
      </section>

      <section aria-label="Friend connections" style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <div>
          <h4 style={{ marginBottom: '0.5rem' }}>Your friends</h4>
          <input
            type="search"
            value={searchTerm}
            onChange={(event) => setSearchTerm(event.target.value)}
            placeholder="Search friends"
            style={{
              width: '100%',
              padding: '0.65rem 0.85rem',
              borderRadius: '0.75rem',
              border: '1px solid rgba(148, 163, 184, 0.25)',
              backgroundColor: 'rgba(15, 23, 42, 0.65)',
              color: '#e2e8f0'
            }}
          />
        </div>
        {filteredConnections.length > 0 ? (
          <ul style={{ listStyle: 'none', padding: 0, display: 'grid', gap: '0.75rem' }}>
            {filteredConnections.map((friend) => (
              <li
                key={friend.id}
                style={{
                  backgroundColor: 'rgba(15, 23, 42, 0.6)',
                  border: '1px solid rgba(148, 163, 184, 0.2)',
                  borderRadius: '0.9rem',
                  padding: '0.9rem 1.1rem',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between'
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                  <span
                    aria-hidden="true"
                    style={{
                      display: 'inline-flex',
                      width: '0.75rem',
                      height: '0.75rem',
                      borderRadius: '9999px',
                      backgroundColor: statusColors[friend.status]
                    }}
                  />
                  <div>
                    <p style={{ marginBottom: '0.15rem', fontWeight: 600 }}>{friend.displayName}</p>
                    <p style={{ margin: 0, color: 'rgba(148, 163, 184, 0.85)', fontSize: '0.8rem' }}>
                      {statusLabels[friend.status]}
                    </p>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        ) : (
          <p style={{ color: 'rgba(148, 163, 184, 0.8)', fontSize: '0.9rem' }}>
            {searchTerm.trim().length > 0
              ? 'No friends match your search yet.'
              : 'Add friends to see their watch activity and status here.'}
          </p>
        )}
      </section>
    </section>
  );
}

export function FriendListManager() {
  const { friends, respondToInvite } = useAppState();
  return <FriendListManagerView friends={friends} respondToInvite={respondToInvite} />;
}
