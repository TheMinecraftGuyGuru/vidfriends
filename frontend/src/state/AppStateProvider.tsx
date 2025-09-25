import {
  ReactNode,
  createContext,
  useCallback,
  useContext,
  useMemo,
  useReducer
} from 'react';

export interface AuthUser {
  id: string;
  email: string;
  displayName: string;
}

export interface FriendInvite {
  id: string;
  displayName: string;
  mutualFriends: number;
}

export interface FriendConnection {
  id: string;
  displayName: string;
  status: 'online' | 'offline' | 'away';
}

export interface FeedEntry {
  id: string;
  title: string;
  sharedBy: string;
  sharedAt: string;
}

interface AppState {
  auth: {
    user: AuthUser | null;
    status: 'authenticated' | 'anonymous';
  };
  friends: {
    pending: FriendInvite[];
    connections: FriendConnection[];
  };
  feed: {
    entries: FeedEntry[];
  };
}

type AppStateAction =
  | { type: 'sign-in'; payload: AuthUser }
  | { type: 'sign-out' }
  | { type: 'add-friend'; payload: FriendConnection }
  | { type: 'add-invite'; payload: FriendInvite }
  | { type: 'resolve-invite'; payload: { inviteId: string; accepted: boolean } }
  | { type: 'share-video'; payload: FeedEntry };

const initialState: AppState = {
  auth: {
    user: null,
    status: 'anonymous'
  },
  friends: {
    pending: [
      { id: 'inv-1', displayName: 'Sasha Rivers', mutualFriends: 4 },
      { id: 'inv-2', displayName: 'Miguel Chen', mutualFriends: 2 }
    ],
    connections: [
      { id: 'friend-1', displayName: 'Rowan Carter', status: 'online' },
      { id: 'friend-2', displayName: 'Priya Das', status: 'away' }
    ]
  },
  feed: {
    entries: [
      {
        id: 'feed-1',
        title: 'Top 10 Cozy Indie Games',
        sharedBy: 'Rowan Carter',
        sharedAt: new Date().toISOString()
      }
    ]
  }
};

function appReducer(state: AppState, action: AppStateAction): AppState {
  switch (action.type) {
    case 'sign-in':
      return {
        ...state,
        auth: {
          user: action.payload,
          status: 'authenticated'
        }
      };
    case 'sign-out':
      return {
        ...state,
        auth: {
          user: null,
          status: 'anonymous'
        }
      };
    case 'add-friend':
      return {
        ...state,
        friends: {
          ...state.friends,
          connections: [action.payload, ...state.friends.connections]
        }
      };
    case 'add-invite':
      return {
        ...state,
        friends: {
          ...state.friends,
          pending: [action.payload, ...state.friends.pending]
        }
      };
    case 'resolve-invite': {
      const invite = state.friends.pending.find((item) => item.id === action.payload.inviteId);
      return {
        ...state,
        friends: {
          pending: state.friends.pending.filter((item) => item.id !== action.payload.inviteId),
          connections:
            action.payload.accepted && invite
              ? [
                  {
                    id: invite.id,
                    displayName: invite.displayName,
                    status: 'offline'
                  },
                  ...state.friends.connections
                ]
              : state.friends.connections
        }
      };
    }
    case 'share-video':
      return {
        ...state,
        feed: {
          entries: [action.payload, ...state.feed.entries]
        }
      };
    default:
      return state;
  }
}

export interface AppStateContextValue extends AppState {
  signIn: (credentials: { email: string; password: string }) => Promise<AuthUser>;
  signUp: (details: { email: string; password: string; displayName: string }) => Promise<AuthUser>;
  requestPasswordReset: (email: string) => Promise<void>;
  signOut: () => void;
  acceptInvite: (inviteId: string) => void;
  declineInvite: (inviteId: string) => void;
  shareVideo: (entry: Pick<FeedEntry, 'title'>) => void;
}

const AppStateContext = createContext<AppStateContextValue | undefined>(undefined);

export function AppStateProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(appReducer, initialState);

  const simulateNetwork = useCallback(async () => {
    await new Promise((resolve) => setTimeout(resolve, 250));
  }, []);

  const signIn = useCallback<AppStateContextValue['signIn']>(
    async ({ email }) => {
      await simulateNetwork();
      const user: AuthUser = {
        id: 'user-1',
        email,
        displayName: email.split('@')[0] ?? 'Friend'
      };
      dispatch({ type: 'sign-in', payload: user });
      return user;
    },
    [simulateNetwork]
  );

  const signUp = useCallback<AppStateContextValue['signUp']>(
    async ({ email, displayName }) => {
      await simulateNetwork();
      const user: AuthUser = {
        id: `user-${Date.now()}`,
        email,
        displayName
      };
      dispatch({ type: 'sign-in', payload: user });
      return user;
    },
    [simulateNetwork]
  );

  const requestPasswordReset = useCallback<AppStateContextValue['requestPasswordReset']>(
    async () => {
      await simulateNetwork();
    },
    [simulateNetwork]
  );

  const signOut = useCallback(() => {
    dispatch({ type: 'sign-out' });
  }, []);

  const acceptInvite = useCallback<AppStateContextValue['acceptInvite']>((inviteId) => {
    dispatch({ type: 'resolve-invite', payload: { inviteId, accepted: true } });
  }, []);

  const declineInvite = useCallback<AppStateContextValue['declineInvite']>((inviteId) => {
    dispatch({ type: 'resolve-invite', payload: { inviteId, accepted: false } });
  }, []);

  const shareVideo = useCallback<AppStateContextValue['shareVideo']>(
    (entry) => {
      dispatch({
        type: 'share-video',
        payload: {
          id: `feed-${Date.now()}`,
          title: entry.title,
          sharedBy: state.auth.user?.displayName ?? 'Anonymous friend',
          sharedAt: new Date().toISOString()
        }
      });
    },
    [state.auth.user?.displayName]
  );

  const contextValue = useMemo<AppStateContextValue>(
    () => ({
      ...state,
      signIn,
      signUp,
      requestPasswordReset,
      signOut,
      acceptInvite,
      declineInvite,
      shareVideo
    }),
    [state, acceptInvite, declineInvite, requestPasswordReset, shareVideo, signIn, signOut, signUp]
  );

  return <AppStateContext.Provider value={contextValue}>{children}</AppStateContext.Provider>;
}

export function useAppStateContext() {
  const value = useContext(AppStateContext);
  if (!value) {
    throw new Error('useAppStateContext must be used within an AppStateProvider');
  }
  return value;
}
