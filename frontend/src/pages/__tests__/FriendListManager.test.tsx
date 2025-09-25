import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import type { FriendConnection, FriendInvite } from '../../state/AppStateProvider';
import { FriendListManagerView } from '../components/FriendListManager';

type FriendCollection = { pending: FriendInvite[]; connections: FriendConnection[] };

const defaultFriends: FriendCollection = {
  pending: [
    { id: 'inv-1', displayName: 'Sasha Rivers', mutualFriends: 4 },
    { id: 'inv-2', displayName: 'Miguel Chen', mutualFriends: 2 }
  ],
  connections: [
    { id: 'friend-1', displayName: 'Rowan Carter', status: 'online' },
    { id: 'friend-2', displayName: 'Priya Das', status: 'away' }
  ]
};

function Harness({
  initialFriends = defaultFriends,
  respondImpl = async () => {}
}: {
  initialFriends?: FriendCollection;
  respondImpl?: (inviteId: string, accepted: boolean) => Promise<void>;
}) {
  const [friends, setFriends] = useState<FriendCollection>(initialFriends);

  const respondToInvite = async (inviteId: string, accepted: boolean) => {
    let invite: FriendInvite | undefined;
    setFriends((prev) => {
      invite = prev.pending.find((item) => item.id === inviteId);
      if (!invite) {
        return prev;
      }
      const nextPending = prev.pending.filter((item) => item.id !== inviteId);
      const nextConnections = accepted
        ? ([{ id: invite.id, displayName: invite.displayName, status: 'offline' as const }, ...prev.connections] as FriendConnection[])
        : prev.connections;
      return {
        pending: nextPending,
        connections: nextConnections
      };
    });

    if (!invite) {
      throw new Error('Invite not found');
    }

    try {
      await respondImpl(inviteId, accepted);
    } catch (error) {
      setFriends((prev) => ({
        pending: [invite as FriendInvite, ...prev.pending],
        connections: accepted
          ? prev.connections.filter((connection) => connection.id !== invite?.id)
          : prev.connections
      }));
      throw error;
    }
  };

  return <FriendListManagerView friends={friends} respondToInvite={respondToInvite} />;
}

describe('FriendListManagerView', () => {
  it('accepts an invite optimistically and surfaces success feedback', async () => {
    const user = userEvent.setup();
    render(<Harness />);

    const acceptButton = screen.getByRole('button', { name: /accept sasha rivers/i });
    await user.click(acceptButton);

    await waitFor(() => expect(screen.getByText(/invitation accepted/i)).toBeInTheDocument());
    await waitFor(() => expect(screen.getByText('1 pending')).toBeInTheDocument());
    expect(screen.getByText(/rowan carter/i)).toBeInTheDocument();
  });

  it('restores the invite and reports the error when the update fails', async () => {
    const respondImpl = vi.fn().mockRejectedValue(new Error('Server unavailable'));
    const user = userEvent.setup();
    render(<Harness respondImpl={respondImpl} />);

    const acceptButton = screen.getByRole('button', { name: /accept sasha rivers/i });
    await user.click(acceptButton);

    await waitFor(() => expect(respondImpl).toHaveBeenCalledWith('inv-1', true));
    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent('Server unavailable')
    );
    expect(acceptButton).not.toBeDisabled();
    expect(screen.getByText('2 pending')).toBeInTheDocument();
  });

  it('filters friends using the search box', async () => {
    const user = userEvent.setup();
    render(<Harness />);

    const searchInput = screen.getByPlaceholderText(/search friends/i);
    await user.type(searchInput, 'rowan');
    expect(screen.getByText(/rowan carter/i)).toBeInTheDocument();

    await user.clear(searchInput);
    await user.type(searchInput, 'zzz');
    expect(screen.getByText(/no friends match your search/i)).toBeInTheDocument();
  });
});
