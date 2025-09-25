import { NavLink, Outlet } from 'react-router-dom';

export function AuthLayout() {
  return (
    <div className="app-shell" style={{ justifyContent: 'center', alignItems: 'center' }}>
      <main
        className="app-main"
        style={{
          maxWidth: '420px',
          width: '100%',
          backgroundColor: 'rgba(15, 23, 42, 0.9)',
          padding: '2rem',
          borderRadius: '1rem',
          boxShadow: '0 30px 50px rgba(15, 23, 42, 0.4)'
        }}
      >
        <NavLink to="/" style={{ display: 'inline-block', marginBottom: '1.5rem' }}>
          ← Back to VidFriends
        </NavLink>
        <Outlet />
      </main>
    </div>
  );
}
