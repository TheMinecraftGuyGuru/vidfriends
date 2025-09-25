import { FormEvent, useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';

import { useAppState } from '../state/useAppState';

export function SignupPage() {
  const navigate = useNavigate();
  const { signUp } = useAppState();
  const [email, setEmail] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [password, setPassword] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSubmitting(true);
    try {
      await signUp({ email, password, displayName });
      navigate('/dashboard');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <section>
      <h2>Create your VidFriends account</h2>
      <form onSubmit={handleSubmit} style={{ display: 'grid', gap: '1rem', marginTop: '1.5rem' }}>
        <label style={{ display: 'grid', gap: '0.5rem' }}>
          Display name
          <input
            required
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
            placeholder="Sam the Streamer"
            style={{ padding: '0.75rem', borderRadius: '0.75rem', border: 'none' }}
          />
        </label>
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
            placeholder="Create a strong password"
            style={{ padding: '0.75rem', borderRadius: '0.75rem', border: 'none' }}
          />
        </label>
        <button
          type="submit"
          disabled={isSubmitting}
          style={{
            background: '#22d3ee',
            color: '#0f172a',
            padding: '0.75rem',
            borderRadius: '0.75rem',
            border: 'none',
            cursor: 'pointer',
            fontWeight: 600
          }}
        >
          {isSubmitting ? 'Creating account...' : 'Create account'}
        </button>
      </form>
      <p style={{ marginTop: '1.5rem' }}>
        Already have an account? <NavLink to="/login">Log in</NavLink>
      </p>
    </section>
  );
}
