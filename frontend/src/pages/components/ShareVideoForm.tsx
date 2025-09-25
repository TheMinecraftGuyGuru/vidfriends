import { FormEvent, useMemo, useState } from 'react';

import { useAppState } from '../../state/useAppState';

export function ShareVideoForm() {
  const { auth, shareVideo } = useAppState();
  const [title, setTitle] = useState('');
  const [url, setUrl] = useState('');
  const [status, setStatus] = useState<'idle' | 'fetching' | 'success' | 'error'>('idle');
  const [feedback, setFeedback] = useState<string | null>(null);

  const isAuthenticated = auth.status === 'authenticated';

  const canSubmit = useMemo(() => {
    if (!isAuthenticated) {
      return false;
    }
    if (status === 'fetching') {
      return false;
    }
    return url.trim().length > 0;
  }, [isAuthenticated, status, url]);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!canSubmit) {
      return;
    }

    setStatus('fetching');
    setFeedback('Fetching video details and saving your share...');

    try {
      const entry = await shareVideo({ title, url });
      setStatus('success');
      setFeedback(`Shared “${entry.title}” with your friends!`);
      setTitle('');
      setUrl('');
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unable to share the video right now. Please try again.';
      setStatus('error');
      setFeedback(message);
    }
  };

  return (
    <section
      aria-label="Share a video"
      style={{
        background: 'rgba(30, 41, 59, 0.55)',
        border: '1px solid rgba(148, 163, 184, 0.25)',
        borderRadius: '1rem',
        padding: '1.5rem',
        display: 'flex',
        flexDirection: 'column',
        gap: '1rem',
        boxShadow: '0 18px 40px rgba(15, 23, 42, 0.35)'
      }}
    >
      <header>
        <h3 style={{ marginBottom: '0.25rem' }}>Share a video</h3>
        <p style={{ color: 'rgba(148, 163, 184, 0.9)', margin: 0 }}>
          Paste a link and we will pull in the title and description once the metadata lookup completes.
        </p>
      </header>

      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <label style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', fontWeight: 600 }}>
          Video URL
          <input
            type="url"
            name="url"
            required
            value={url}
            onChange={(event) => setUrl(event.target.value)}
            placeholder="https://www.youtube.com/watch?v=example"
            style={{
              padding: '0.65rem 0.75rem',
              borderRadius: '0.75rem',
              border: '1px solid rgba(148, 163, 184, 0.35)',
              background: 'rgba(15, 23, 42, 0.7)',
              color: '#f8fafc'
            }}
          />
        </label>

        <label style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', fontWeight: 600 }}>
          Optional title override
          <input
            type="text"
            name="title"
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            placeholder="Add a note or custom title"
            style={{
              padding: '0.65rem 0.75rem',
              borderRadius: '0.75rem',
              border: '1px solid rgba(148, 163, 184, 0.25)',
              background: 'rgba(15, 23, 42, 0.7)',
              color: '#f8fafc'
            }}
          />
        </label>

        <button
          type="submit"
          disabled={!canSubmit}
          style={{
            marginTop: '0.5rem',
            padding: '0.75rem 1rem',
            borderRadius: '0.75rem',
            border: 'none',
            background: canSubmit ? 'linear-gradient(135deg, #6366f1, #8b5cf6)' : 'rgba(99, 102, 241, 0.35)',
            color: '#f8fafc',
            fontWeight: 600,
            cursor: canSubmit ? 'pointer' : 'not-allowed',
            transition: 'transform 0.15s ease, box-shadow 0.15s ease',
            boxShadow: canSubmit ? '0 12px 30px rgba(79, 70, 229, 0.35)' : 'none'
          }}
        >
          {status === 'fetching' ? 'Sharing…' : 'Share video'}
        </button>
      </form>

      <div aria-live="polite" style={{ minHeight: '1.5rem', color: '#f1f5f9' }}>
        {!isAuthenticated && (
          <p style={{ margin: 0, color: 'rgba(148, 163, 184, 0.85)' }}>
            Sign in to start sharing videos with your friends.
          </p>
        )}
        {status === 'fetching' && (
          <p style={{ margin: 0, color: 'rgba(148, 163, 184, 0.95)' }}>{feedback}</p>
        )}
        {status === 'success' && feedback && (
          <p style={{ margin: 0, color: 'rgba(129, 199, 132, 0.95)' }}>{feedback}</p>
        )}
        {status === 'error' && feedback && (
          <p style={{ margin: 0, color: 'rgba(248, 113, 113, 0.95)' }}>{feedback}</p>
        )}
      </div>
    </section>
  );
}
