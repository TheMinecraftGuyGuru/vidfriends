import { NavLink } from 'react-router-dom';

import { useAppState } from '../state/useAppState';

export function HomePage() {
  const { auth } = useAppState();

  return (
    <section>
      <header>
        <h2>Welcome to VidFriends</h2>
        <p>
          Connect with friends, share the latest videos, and see what everyone is
          watching in real time.
        </p>
      </header>
      <div style={{ marginTop: '2rem', display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
        {!auth.user && (
          <NavLink
            to="/signup"
            style={{
              background: '#38bdf8',
              color: '#0f172a',
              padding: '0.75rem 1.5rem',
              borderRadius: '999px',
              fontWeight: 600
            }}
          >
            Get Started
          </NavLink>
        )}
        <NavLink
          to="/dashboard"
          style={{
            border: '1px solid rgba(148, 163, 184, 0.5)',
            padding: '0.75rem 1.5rem',
            borderRadius: '999px'
          }}
        >
          View Dashboard
        </NavLink>
      </div>
    </section>
  );
}
