import { FormEvent, useState } from 'react';

import { useAppState } from '../state/useAppState';

export function ForgotPasswordPage() {
  const { requestPasswordReset } = useAppState();
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState<'idle' | 'submitting' | 'sent'>('idle');

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setStatus('submitting');
    await requestPasswordReset(email);
    setStatus('sent');
  };

  return (
    <section>
      <h2>Reset your password</h2>
      <p style={{ color: 'rgba(148, 163, 184, 0.9)' }}>
        Enter the email associated with your VidFriends account to receive a reset link.
      </p>
      <form onSubmit={handleSubmit} style={{ display: 'grid', gap: '1rem', marginTop: '1.5rem' }}>
        <label style={{ display: 'grid', gap: '0.5rem' }}>
          Email address
          <input
            required
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="you@example.com"
            style={{ padding: '0.75rem', borderRadius: '0.75rem', border: 'none' }}
          />
        </label>
        <button
          type="submit"
          disabled={status === 'submitting'}
          style={{
            background: '#c084fc',
            color: '#0f172a',
            padding: '0.75rem',
            borderRadius: '0.75rem',
            border: 'none',
            cursor: 'pointer',
            fontWeight: 600
          }}
        >
          {status === 'submitting' ? 'Sending link...' : 'Send reset link'}
        </button>
      </form>
      {status === 'sent' && (
        <p style={{ marginTop: '1rem', color: '#a7f3d0' }}>
          If an account exists for <strong>{email}</strong> you&apos;ll receive a reset link shortly.
        </p>
      )}
    </section>
  );
}
