import { useMemo } from 'react';

import { useAppStateContext } from './AppStateProvider';

export function useAppState() {
  const context = useAppStateContext();

  return useMemo(
    () => ({
      auth: context.auth,
      friends: context.friends,
      feed: context.feed,
      signIn: context.signIn,
      signUp: context.signUp,
      signOut: context.signOut,
      shareVideo: context.shareVideo,
      acceptInvite: context.acceptInvite,
      declineInvite: context.declineInvite,
      requestPasswordReset: context.requestPasswordReset
    }),
    [context]
  );
}
