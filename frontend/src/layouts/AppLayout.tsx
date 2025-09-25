import { NavLink, Outlet } from 'react-router-dom';

import { useAppState } from '../state/useAppState';

export function AppLayout() {
  const { auth, signOut } = useAppState();

  return (
    <div className="app-shell">
      <aside style={{ width: '220px', padding: '2rem', backgroundColor: '#0b1120' }}>
        <h1 style={{ fontSize: '1.5rem', marginBottom: '2rem' }}>VidFriends</h1>
        <nav style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          <NavLink to="/" end>
            Home
          </NavLink>
          <NavLink to="/dashboard">Dashboard</NavLink>
          {!auth.user && <NavLink to="/login">Log in</NavLink>}
        </nav>
        {auth.user && (
          <div style={{ marginTop: '2rem' }}>
            <p style={{ fontSize: '0.9rem', marginBottom: '0.5rem' }}>
              Signed in as <strong>{auth.user.displayName}</strong>
            </p>
            <button type="button" onClick={signOut} style={{ cursor: 'pointer' }}>
              Sign out
            </button>
          </div>
        )}
      </aside>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
