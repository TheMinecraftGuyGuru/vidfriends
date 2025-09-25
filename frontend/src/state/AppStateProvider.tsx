import {
  ReactNode,
  createContext,
  useCallback,
  useContext,
  useMemo,
  useReducer
} from 'react';

const SESSION_STORAGE_KEY = 'vidfriends.session';
const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');

class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function parseJSON(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return null;
  }
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}

function getErrorMessage(payload: unknown): string | null {
  if (payload && typeof payload === 'object' && 'error' in payload) {
    const message = (payload as { error?: unknown }).error;
    if (typeof message === 'string' && message.trim().length > 0) {
      return message;
    }
  }
  return null;
}

function buildURL(path: string): string {
  if (!path.startsWith('/')) {
    return `${API_BASE_URL}/${path}`;
  }
  return `${API_BASE_URL}${path}`;
}

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  let response: Response;
  try {
    response = await fetch(buildURL(path), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    });
  } catch {
    throw new Error('Unable to connect to VidFriends services. Please try again.');
  }

  const payload = await parseJSON(response);
  if (!response.ok) {
    const message = getErrorMessage(payload) ?? `Request failed with status ${response.status}`;
    throw new ApiError(message, response.status);
  }

  return (payload as T) ?? (undefined as T);
}

function normalizeEmail(email: string): string {
  return email.trim().toLowerCase();
}

function deriveDisplayName(email: string, provided?: string): string {
  if (provided && provided.trim().length > 0) {
    return provided.trim();
  }
  const username = email.split('@')[0];
  return username && username.length > 0 ? username : 'Friend';
}

interface StoredSession {
  user: AuthUser;
  tokens: SessionTokens;
}

function loadStoredSession(): StoredSession | null {
  if (typeof window === 'undefined') {
    return null;
  }
  const raw = window.localStorage.getItem(SESSION_STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    const parsed = JSON.parse(raw) as StoredSession;
    if (!parsed || typeof parsed !== 'object') {
      return null;
    }
    if (!parsed.user || !parsed.tokens) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function persistSession(session: StoredSession) {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
}

function clearStoredSession() {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.removeItem(SESSION_STORAGE_KEY);
}

interface SessionTokens {
  accessToken: string;
  accessExpiresAt: string;
  refreshToken: string;
  refreshExpiresAt: string;
}

interface AuthResponse {
  tokens: SessionTokens;
}

function ensureTokens(response: AuthResponse): SessionTokens {
  if (!response.tokens) {
    throw new Error('Authentication response was missing session tokens.');
  }
  return {
    accessToken: response.tokens.accessToken,
    accessExpiresAt: response.tokens.accessExpiresAt,
    refreshToken: response.tokens.refreshToken,
    refreshExpiresAt: response.tokens.refreshExpiresAt
  };
}

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

interface AuthState {
  user: AuthUser | null;
  status: 'authenticated' | 'anonymous';
  tokens: SessionTokens | null;
}

interface AppState {
  auth: AuthState;
  friends: {
    pending: FriendInvite[];
    connections: FriendConnection[];
  };
  feed: {
    entries: FeedEntry[];
  };
}

type AppStateAction =
  | { type: 'sign-in'; payload: { user: AuthUser; tokens: SessionTokens } }
  | { type: 'sign-out' }
  | { type: 'add-friend'; payload: FriendConnection }
  | { type: 'add-invite'; payload: FriendInvite }
  | { type: 'resolve-invite'; payload: { inviteId: string; accepted: boolean } }
  | { type: 'remove-friend'; payload: { friendId: string } }
  | { type: 'share-video'; payload: FeedEntry };

const initialState: AppState = {
  auth: {
    user: null,
    status: 'anonymous',
    tokens: null
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

function initializeState(defaultState: AppState): AppState {
  const stored = loadStoredSession();
  if (!stored) {
    return defaultState;
  }
  return {
    ...defaultState,
    auth: {
      user: stored.user,
      status: 'authenticated',
      tokens: stored.tokens
    }
  };
}

function appReducer(state: AppState, action: AppStateAction): AppState {
  switch (action.type) {
    case 'sign-in':
      return {
        ...state,
        auth: {
          user: action.payload.user,
          status: 'authenticated',
          tokens: action.payload.tokens
        }
      };
    case 'sign-out':
      return {
        ...state,
        auth: {
          user: null,
          status: 'anonymous',
          tokens: null
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
    case 'remove-friend':
      return {
        ...state,
        friends: {
          ...state.friends,
          connections: state.friends.connections.filter((friend) => friend.id !== action.payload.friendId)
        }
      };
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
  respondToInvite: (inviteId: string, accepted: boolean) => Promise<void>;
  shareVideo: (entry: Pick<FeedEntry, 'title'>) => void;
}

const AppStateContext = createContext<AppStateContextValue | undefined>(undefined);

export function AppStateProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(appReducer, initialState, initializeState);

  const signIn = useCallback<AppStateContextValue['signIn']>(
    async ({ email, password }) => {
      const normalizedEmail = normalizeEmail(email);
      try {
        const response = await postJSON<AuthResponse>('/api/v1/auth/login', {
          email: normalizedEmail,
          password
        });
        const tokens = ensureTokens(response);
        const user: AuthUser = {
          id: normalizedEmail,
          email: normalizedEmail,
          displayName: deriveDisplayName(normalizedEmail)
        };
        dispatch({ type: 'sign-in', payload: { user, tokens } });
        persistSession({ user, tokens });
        return user;
      } catch (error) {
        if (error instanceof ApiError) {
          throw new Error(error.message);
        }
        throw error instanceof Error ? error : new Error('Unable to sign in. Please try again.');
      }
    },
    [dispatch]
  );

  const signUp = useCallback<AppStateContextValue['signUp']>(
    async ({ email, password, displayName }) => {
      const normalizedEmail = normalizeEmail(email);
      try {
        const response = await postJSON<AuthResponse>('/api/v1/auth/signup', {
          email: normalizedEmail,
          password
        });
        const tokens = ensureTokens(response);
        const user: AuthUser = {
          id: normalizedEmail,
          email: normalizedEmail,
          displayName: deriveDisplayName(normalizedEmail, displayName)
        };
        dispatch({ type: 'sign-in', payload: { user, tokens } });
        persistSession({ user, tokens });
        return user;
      } catch (error) {
        if (error instanceof ApiError) {
          throw new Error(error.message);
        }
        throw error instanceof Error ? error : new Error('Unable to create your account. Please try again.');
      }
    },
    [dispatch]
  );

  const requestPasswordReset = useCallback<AppStateContextValue['requestPasswordReset']>(
    async (email: string) => {
      const normalizedEmail = normalizeEmail(email);
      if (!normalizedEmail) {
        throw new Error('Please provide the email associated with your account.');
      }
      try {
        await postJSON('/api/v1/auth/password-reset', { email: normalizedEmail });
      } catch (error) {
        if (error instanceof ApiError) {
          if (error.status === 404) {
            // Password reset API is not yet available; treat as success so the user can continue.
            return;
          }
          throw new Error(error.message);
        }
        throw error instanceof Error
          ? error
          : new Error('Unable to request a password reset at this time. Please try again.');
      }
    },
    []
  );

  const signOut = useCallback(() => {
    clearStoredSession();
    dispatch({ type: 'sign-out' });
  }, []);

  const acceptInvite = useCallback<AppStateContextValue['acceptInvite']>((inviteId) => {
    dispatch({ type: 'resolve-invite', payload: { inviteId, accepted: true } });
  }, []);

  const declineInvite = useCallback<AppStateContextValue['declineInvite']>((inviteId) => {
    dispatch({ type: 'resolve-invite', payload: { inviteId, accepted: false } });
  }, []);

  const respondToInvite = useCallback<AppStateContextValue['respondToInvite']>(
    async (inviteId, accepted) => {
      const invite = state.friends.pending.find((item) => item.id === inviteId);
      if (!invite) {
        throw new Error('This invitation is no longer available.');
      }

      dispatch({ type: 'resolve-invite', payload: { inviteId, accepted } });

      try {
        if (API_BASE_URL) {
          await postJSON('/api/v1/friends/invitations/respond', {
            inviteId,
            accepted
          });
        } else {
          await new Promise((resolve) => {
            setTimeout(resolve, 400);
          });
        }
      } catch (error) {
        if (accepted) {
          dispatch({ type: 'remove-friend', payload: { friendId: inviteId } });
        }
        dispatch({ type: 'add-invite', payload: invite });

        if (error instanceof ApiError) {
          throw new Error(error.message);
        }
        throw error instanceof Error
          ? error
          : new Error('Unable to update the invitation at this time. Please try again.');
      }
    },
    [state.friends.pending]
  );

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
      respondToInvite,
      shareVideo
    }),
    [
      state,
      acceptInvite,
      declineInvite,
      requestPasswordReset,
      respondToInvite,
      shareVideo,
      signIn,
      signOut,
      signUp
    ]
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
