import { FormEvent, useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';

import { useAppState } from '../state/useAppState';

export function LoginPage() {
  const navigate = useNavigate();
  const { signIn } = useAppState();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    try {
      await signIn({ email, password });
      navigate('/dashboard');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to sign in');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <section>
      <h2>Log in</h2>
      <p style={{ color: 'rgba(148, 163, 184, 0.9)' }}>
        Access your VidFriends account to continue sharing videos.
      </p>
      <form onSubmit={handleSubmit} style={{ display: 'grid', gap: '1rem', marginTop: '1.5rem' }}>
        <label style={{ display: 'grid', gap: '0.5rem' }}>
          Email
          <input
            required
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="you@example.com"
            style={{ padding: '0.75rem', borderRadius: '0.75rem', border: 'none' }}
          />
        </label>
        <label style={{ display: 'grid', gap: '0.5rem' }}>
          Password
          <input
            required
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder="••••••••"
            style={{ padding: '0.75rem', borderRadius: '0.75rem', border: 'none' }}
          />
        </label>
        {error && <p style={{ color: '#fda4af' }}>{error}</p>}
        <button
          type="submit"
          disabled={isSubmitting}
          style={{
            background: '#38bdf8',
            color: '#0f172a',
            padding: '0.75rem',
            borderRadius: '0.75rem',
            border: 'none',
            cursor: 'pointer',
            fontWeight: 600
          }}
        >
          {isSubmitting ? 'Signing in...' : 'Log in'}
        </button>
      </form>
      <div style={{ marginTop: '1.5rem', display: 'grid', gap: '0.5rem' }}>
        <NavLink to="/forgot-password">Forgot your password?</NavLink>
        <NavLink to="/signup">Need an account? Sign up</NavLink>
      </div>
    </section>
  );
}
