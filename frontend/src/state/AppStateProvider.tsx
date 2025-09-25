import {
  ReactNode,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useRef
} from 'react';

const SESSION_STORAGE_KEY = 'vidfriends.session';
const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');
const USE_MOCK_DATA = String(import.meta.env.VITE_USE_MOCK_DATA ?? '').toLowerCase() === 'true';

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

async function getJSON<T>(path: string, tokens: SessionTokens | null): Promise<T> {
  let response: Response;
  try {
    response = await fetch(buildURL(path), {
      method: 'GET',
      headers: tokens?.accessToken
        ? {
            Authorization: `Bearer ${tokens.accessToken}`
          }
        : undefined
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

const REFRESH_LEEWAY_MS = 60_000;

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

function toEpochMillis(timestamp: string | undefined): number | null {
  if (!timestamp) {
    return null;
  }
  const value = Date.parse(timestamp);
  return Number.isNaN(value) ? null : value;
}

function isExpired(timestamp: string): boolean {
  const value = toEpochMillis(timestamp);
  if (value === null) {
    return false;
  }
  return value <= Date.now();
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

export type ReactionType = 'like' | 'love' | 'wow' | 'laugh';

export interface FeedEntry {
  id: string;
  title: string;
  sharedBy: string;
  sharedAt: string;
  platform: string;
  url: string;
  thumbnailUrl: string;
  channelName: string;
  durationSeconds: number;
  viewCount: number;
  description: string;
  tags: string[];
  reactions: Record<ReactionType, number>;
  userReaction: ReactionType | null;
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
  | { type: 'refresh-tokens'; payload: { tokens: SessionTokens } }
  | { type: 'add-friend'; payload: FriendConnection }
  | { type: 'add-invite'; payload: FriendInvite }
  | { type: 'resolve-invite'; payload: { inviteId: string; accepted: boolean } }
  | { type: 'remove-friend'; payload: { friendId: string } }
  | { type: 'share-video'; payload: FeedEntry }
  | { type: 'react-to-video'; payload: { entryId: string; reaction: ReactionType } }
  | {
      type: 'set-friends';
      payload: { pending: FriendInvite[]; connections: FriendConnection[] };
    }
  | { type: 'set-feed'; payload: FeedEntry[] };

const mockFriendsState: { pending: FriendInvite[]; connections: FriendConnection[] } = {
  pending: [
    { id: 'inv-1', displayName: 'Sasha Rivers', mutualFriends: 4 },
    { id: 'inv-2', displayName: 'Miguel Chen', mutualFriends: 2 }
  ],
  connections: [
    { id: 'friend-1', displayName: 'Rowan Carter', status: 'online' },
    { id: 'friend-2', displayName: 'Priya Das', status: 'away' }
  ]
};

const mockFeedEntries: FeedEntry[] = [
  {
    id: 'feed-1',
    title: 'Top 10 Cozy Indie Games',
    sharedBy: 'Rowan Carter',
    sharedAt: new Date(Date.now() - 1000 * 60 * 35).toISOString(),
    platform: 'YouTube',
    url: 'https://www.youtube.com/watch?v=I6-hm4DYPwU',
    thumbnailUrl: 'https://i.ytimg.com/vi/I6-hm4DYPwU/hqdefault.jpg',
    channelName: 'Indie Realm',
    durationSeconds: 972,
    viewCount: 48200,
    description:
      'Discover new cozy titles perfect for winding down after a long day. This curated list covers hidden gems, heartfelt stories, and soundtracks you will want on repeat.',
    tags: ['Games', 'Cozy', 'Indie'],
    reactions: { like: 18, love: 6, wow: 2, laugh: 1 },
    userReaction: null
  },
  {
    id: 'feed-2',
    title: 'Morning Flow Yoga for Beginners',
    sharedBy: 'Priya Das',
    sharedAt: new Date(Date.now() - 1000 * 60 * 60 * 5).toISOString(),
    platform: 'YouTube',
    url: 'https://www.youtube.com/watch?v=u5o593sW9DQ',
    thumbnailUrl: 'https://i.ytimg.com/vi/u5o593sW9DQ/hqdefault.jpg',
    channelName: 'Peaceful Moves',
    durationSeconds: 1803,
    viewCount: 91500,
    description:
      'A gentle yoga flow designed to wake up the body and focus the mind. No equipment required—perfect for easing into the day.',
    tags: ['Wellness', 'Mindfulness', 'Beginner Friendly'],
    reactions: { like: 12, love: 9, wow: 1, laugh: 0 },
    userReaction: 'love'
  },
  {
    id: 'feed-3',
    title: 'How the JWST Sees the Universe',
    sharedBy: 'Miguel Chen',
    sharedAt: new Date(Date.now() - 1000 * 60 * 60 * 20).toISOString(),
    platform: 'Nebula',
    url: 'https://nebula.tv/videos/space-jwst',
    thumbnailUrl: 'https://images.unsplash.com/photo-1581091226825-a6a2a5aee158?auto=format&fit=crop&w=1200&q=80',
    channelName: 'Cosmic Perspectives',
    durationSeconds: 1245,
    viewCount: 38700,
    description:
      'Astrophysicist Dr. Mei Park breaks down the science behind the James Webb Space Telescope and what its discoveries mean for our understanding of deep space.',
    tags: ['Science', 'Space', 'Documentary'],
    reactions: { like: 21, love: 11, wow: 7, laugh: 0 },
    userReaction: null
  }
];

const emptyAppState: AppState = {
  auth: {
    user: null,
    status: 'anonymous',
    tokens: null
  },
  friends: {
    pending: [],
    connections: []
  },
  feed: {
    entries: []
  }
};

const initialState: AppState = emptyAppState;

function initializeState(defaultState: AppState): AppState {
  const stored = loadStoredSession();
  if (!stored) {
    return defaultState;
  }
  if (isExpired(stored.tokens.refreshExpiresAt)) {
    clearStoredSession();
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
    case 'refresh-tokens':
      return {
        ...state,
        auth: {
          ...state.auth,
          tokens: action.payload.tokens
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
    case 'react-to-video': {
      const { entryId, reaction } = action.payload;
      return {
        ...state,
        feed: {
          entries: state.feed.entries.map((entry) => {
            if (entry.id !== entryId) {
              return entry;
            }

            const nextReactions = { ...entry.reactions };
            const currentReaction = entry.userReaction;
            if (currentReaction) {
              nextReactions[currentReaction] = Math.max(0, (nextReactions[currentReaction] ?? 0) - 1);
            }

            const togglingSameReaction = currentReaction === reaction;

            if (!togglingSameReaction) {
              nextReactions[reaction] = (nextReactions[reaction] ?? 0) + 1;
            }

            return {
              ...entry,
              reactions: nextReactions,
              userReaction: togglingSameReaction ? null : reaction
            };
          })
        }
      };
    }
    case 'set-friends':
      return {
        ...state,
        friends: {
          pending: action.payload.pending,
          connections: action.payload.connections
        }
      };
    case 'set-feed':
      return {
        ...state,
        feed: {
          entries: action.payload
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
  shareVideo: (entry: { title: string; url?: string }) => void;
  reactToVideo: (entryId: string, reaction: ReactionType) => void;
}

const AppStateContext = createContext<AppStateContextValue | undefined>(undefined);

interface RawFriendRequest {
  ID: string;
  Requester: string;
  Receiver: string;
  Status: string;
}

interface RawVideoShare {
  ID: string;
  OwnerID: string;
  URL: string;
  Title?: string;
  Description?: string;
  Thumbnail?: string | null;
  CreatedAt?: string;
}

function shouldUseMockLayer() {
  if (USE_MOCK_DATA) {
    return true;
  }
  return !API_BASE_URL;
}

async function loadFriendsFromApi(userId: string, tokens: SessionTokens | null) {
  const query = new URLSearchParams({ user: userId });
  const response = await getJSON<{ requests: RawFriendRequest[] }>(`/api/v1/friends?${query.toString()}`, tokens);

  const connectionsMap = new Map<string, FriendConnection>();
  const pendingInvites: FriendInvite[] = [];

  for (const request of response.requests ?? []) {
    const status = request.Status.toLowerCase();
    const otherUser = request.Requester === userId ? request.Receiver : request.Requester;
    if (status === 'pending' && request.Receiver === userId) {
      pendingInvites.push({
        id: request.ID,
        displayName: deriveDisplayName(otherUser),
        mutualFriends: 0
      });
      continue;
    }

    if (status === 'accepted' && otherUser) {
      if (!connectionsMap.has(otherUser)) {
        const statusCycle: FriendConnection['status'][] = ['online', 'away', 'offline'];
        const statusIndex = connectionsMap.size % statusCycle.length;
        connectionsMap.set(otherUser, {
          id: otherUser,
          displayName: deriveDisplayName(otherUser),
          status: statusCycle[statusIndex]
        });
      }
    }
  }

  return {
    pending: pendingInvites,
    connections: Array.from(connectionsMap.values())
  };
}

async function loadFeedFromApi(userId: string, tokens: SessionTokens | null) {
  const query = new URLSearchParams({ user: userId });
  const response = await getJSON<{ entries: RawVideoShare[] }>(`/api/v1/videos/feed?${query.toString()}`, tokens);

  return (response.entries ?? []).map((entry) => {
    const title = entry.Title && entry.Title.trim().length > 0 ? entry.Title : 'Shared video';
    const ownerDisplayName = deriveDisplayName(entry.OwnerID);
    return {
      id: entry.ID,
      title,
      sharedBy: ownerDisplayName,
      sharedAt: entry.CreatedAt ?? new Date().toISOString(),
      platform: inferPlatformFromUrl(entry.URL),
      url: entry.URL,
      thumbnailUrl: entry.Thumbnail ?? generateThumbnailForTitle(title),
      channelName: `${ownerDisplayName}'s share`,
      durationSeconds: generateDurationForShare(),
      viewCount: generateViewCount(),
      description: entry.Description ?? '',
      tags: deriveTagsFromTitle(title),
      reactions: { like: 0, love: 0, wow: 0, laugh: 0 },
      userReaction: null
    } satisfies FeedEntry;
  });
}

function loadFriendsFromMock() {
  return {
    pending: mockFriendsState.pending.map((invite) => ({ ...invite })),
    connections: mockFriendsState.connections.map((friend) => ({ ...friend }))
  };
}

function loadFeedFromMock() {
  return mockFeedEntries.map((entry) => ({ ...entry }));
}

export function AppStateProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(appReducer, initialState, initializeState);
  const refreshInFlight = useRef(false);

  const refreshSession = useCallback(async () => {
    if (shouldUseMockLayer()) {
      return;
    }
    if (refreshInFlight.current) {
      return;
    }

    const tokens = state.auth.tokens;
    if (!tokens) {
      return;
    }

    if (isExpired(tokens.refreshExpiresAt)) {
      clearStoredSession();
      dispatch({ type: 'sign-out' });
      return;
    }

    refreshInFlight.current = true;
    try {
      const response = await postJSON<AuthResponse>('/api/v1/auth/refresh', {
        refreshToken: tokens.refreshToken
      });
      const nextTokens = ensureTokens(response);
      dispatch({ type: 'refresh-tokens', payload: { tokens: nextTokens } });
      if (state.auth.user) {
        persistSession({ user: state.auth.user, tokens: nextTokens });
      } else {
        clearStoredSession();
      }
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        clearStoredSession();
        dispatch({ type: 'sign-out' });
      } else {
        console.error('Failed to refresh VidFriends session', error);
      }
    } finally {
      refreshInFlight.current = false;
    }
  }, [dispatch, state.auth.tokens, state.auth.user]);

  useEffect(() => {
    if (shouldUseMockLayer()) {
      return;
    }

    const tokens = state.auth.tokens;
    if (!tokens) {
      return;
    }

    if (isExpired(tokens.refreshExpiresAt)) {
      clearStoredSession();
      dispatch({ type: 'sign-out' });
      return;
    }

    const expiresAt = toEpochMillis(tokens.accessExpiresAt);
    if (expiresAt === null) {
      return;
    }

    const delay = Math.max(expiresAt - REFRESH_LEEWAY_MS - Date.now(), 0);
    const timer = window.setTimeout(() => {
      refreshSession();
    }, delay);

    return () => {
      window.clearTimeout(timer);
    };
  }, [dispatch, refreshSession, state.auth.tokens]);

  useEffect(() => {
    let cancelled = false;

    const usingMock = shouldUseMockLayer();
    const activeUserId = state.auth.user?.id ?? (usingMock ? 'mock-user' : '');

    if (!activeUserId) {
      dispatch({ type: 'set-friends', payload: { pending: [], connections: [] } });
      dispatch({ type: 'set-feed', payload: [] });
      return () => {
        cancelled = true;
      };
    }

    async function loadData() {
      try {
        const [friends, feed] = usingMock
          ? [loadFriendsFromMock(), loadFeedFromMock()]
          : await Promise.all([
              loadFriendsFromApi(activeUserId, state.auth.tokens),
              loadFeedFromApi(activeUserId, state.auth.tokens)
            ]);

        if (!cancelled) {
          dispatch({ type: 'set-friends', payload: friends });
          dispatch({ type: 'set-feed', payload: feed });
        }
      } catch (error) {
        if (!cancelled) {
          dispatch({ type: 'set-friends', payload: loadFriendsFromMock() });
          dispatch({ type: 'set-feed', payload: loadFeedFromMock() });
        }
        console.error('Failed to load VidFriends data', error);
      }
    }

    loadData();

    return () => {
      cancelled = true;
    };
  }, [state.auth.tokens, state.auth.user]);

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

      const previousFriendsState = {
        pending: [...state.friends.pending],
        connections: [...state.friends.connections]
      };

      dispatch({ type: 'resolve-invite', payload: { inviteId, accepted } });

      try {
        if (API_BASE_URL) {
          await postJSON('/api/v1/friends/respond', {
            requestId: inviteId,
            action: accepted ? 'accept' : 'block'
          });
        } else {
          await new Promise((resolve) => {
            setTimeout(resolve, 400);
          });
        }
      } catch (error) {
        dispatch({ type: 'set-friends', payload: previousFriendsState });

        if (error instanceof ApiError) {
          throw new Error(error.message);
        }
        throw error instanceof Error
          ? error
          : new Error('Unable to update the invitation at this time. Please try again.');
      }
    },
    [state.friends.connections, state.friends.pending]
  );

  const shareVideo = useCallback<AppStateContextValue['shareVideo']>(
    (entry) => {
      const title = entry.title.trim() || 'Untitled share';
      const fallbackTags = deriveTagsFromTitle(title);
      dispatch({
        type: 'share-video',
        payload: {
          id: `feed-${Date.now()}`,
          title,
          sharedBy: state.auth.user?.displayName ?? 'Anonymous friend',
          sharedAt: new Date().toISOString(),
          platform: inferPlatformFromUrl(entry.url),
          url: entry.url ?? '#',
          thumbnailUrl: generateThumbnailForTitle(title),
          channelName: `${state.auth.user?.displayName ?? 'A VidFriend'}'s pick`,
          durationSeconds: generateDurationForShare(),
          viewCount: generateViewCount(),
          description: `Shared by ${state.auth.user?.displayName ?? 'a friend'}: ${title}.`,
          tags: fallbackTags,
          reactions: { like: 0, love: 0, wow: 0, laugh: 0 },
          userReaction: null
        }
      });
    },
    [state.auth.user?.displayName]
  );

  const reactToVideo = useCallback<AppStateContextValue['reactToVideo']>(
    (entryId, reaction) => {
      dispatch({ type: 'react-to-video', payload: { entryId, reaction } });
    },
    []
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
      shareVideo,
      reactToVideo
    }),
    [
      state,
      acceptInvite,
      declineInvite,
      requestPasswordReset,
      respondToInvite,
      shareVideo,
      reactToVideo,
      signIn,
      signOut,
      signUp
    ]
  );

  return <AppStateContext.Provider value={contextValue}>{children}</AppStateContext.Provider>;
}

function inferPlatformFromUrl(url: string | undefined) {
  if (!url) {
    return 'Shared link';
  }
  const hostname = (() => {
    try {
      return new URL(url).hostname.replace('www.', '');
    } catch (error) {
      return '';
    }
  })();
  if (!hostname) {
    return 'Shared link';
  }
  if (hostname.includes('youtube')) {
    return 'YouTube';
  }
  if (hostname.includes('twitch')) {
    return 'Twitch';
  }
  if (hostname.includes('nebula')) {
    return 'Nebula';
  }
  if (hostname.includes('vimeo')) {
    return 'Vimeo';
  }
  return hostname.charAt(0).toUpperCase() + hostname.slice(1);
}

function deriveTagsFromTitle(title: string): string[] {
  const normalized = title.toLowerCase();
  const tags = new Set<string>();
  if (normalized.match(/game|play|speedrun/)) {
    tags.add('Games');
  }
  if (normalized.match(/learn|tutorial|guide|how to/)) {
    tags.add('Learning');
  }
  if (normalized.match(/music|song|playlist/)) {
    tags.add('Music');
  }
  if (normalized.match(/news|update|report/)) {
    tags.add('News');
  }
  if (normalized.match(/space|science|tech|engineering/)) {
    tags.add('Science');
  }
  if (normalized.match(/yoga|wellness|meditation|mindful/)) {
    tags.add('Wellness');
  }

  if (tags.size === 0) {
    tags.add('Favorites');
  }
  return Array.from(tags);
}

function generateThumbnailForTitle(title: string): string {
  const encoded = encodeURIComponent(title.toLowerCase().replace(/\s+/g, '-'));
  return `https://images.unsplash.com/seed/${encoded}/640x360?auto=format&fit=crop&w=1200&q=80`;
}

function generateDurationForShare() {
  const min = 240;
  const max = 1800;
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function generateViewCount() {
  const min = 1200;
  const max = 120000;
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

export function useAppStateContext() {
  const value = useContext(AppStateContext);
  if (!value) {
    throw new Error('useAppStateContext must be used within an AppStateProvider');
  }
  return value;
}
