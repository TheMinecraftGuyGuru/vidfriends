import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { act, useEffect } from 'react';
import { afterEach, vi } from 'vitest';

import { ToastProvider, useToast } from '../ToastProvider';

function ToastTrigger({
  message,
  options
}: {
  message: string;
  options?: { variant?: 'info' | 'success' | 'error'; duration?: number };
}) {
  const { showToast } = useToast();
  return (
    <button type="button" onClick={() => showToast(message, options)}>
      Show toast
    </button>
  );
}

function ToastRegistrar({
  message,
  options,
  onReady
}: {
  message: string;
  options?: { variant?: 'info' | 'success' | 'error'; duration?: number };
  onReady: (trigger: () => void) => void;
}) {
  const { showToast } = useToast();
  useEffect(() => {
    onReady(() => {
      showToast(message, options);
    });
  }, [message, onReady, options, showToast]);

  return null;
}

afterEach(() => {
  vi.useRealTimers();
});

describe('ToastProvider', () => {
  it('renders a toast with the default info variant', async () => {
    const user = userEvent.setup();

    render(
      <ToastProvider>
        <ToastTrigger message="Informational" />
      </ToastProvider>
    );

    await user.click(screen.getByRole('button', { name: /show toast/i }));

    const toast = screen.getByText('Informational').closest('.toast');
    expect(toast).not.toBeNull();
    expect(toast).toHaveClass('toast-info');
  });

  it('allows a toast to be dismissed manually', async () => {
    const user = userEvent.setup();

    render(
      <ToastProvider>
        <ToastTrigger message="Dismiss me" />
      </ToastProvider>
    );

    await user.click(screen.getByRole('button', { name: /show toast/i }));

    const dismissButton = screen.getByRole('button', { name: /dismiss notification/i });
    await user.click(dismissButton);

    expect(screen.queryByText('Dismiss me')).not.toBeInTheDocument();
  });

  it('clears toasts automatically after the provided duration', () => {
    vi.useFakeTimers();
    let triggerToast: (() => void) | undefined;

    render(
      <ToastProvider>
        <ToastRegistrar
          message="Ephemeral"
          options={{ duration: 2000 }}
          onReady={(trigger) => {
            triggerToast = trigger;
          }}
        />
      </ToastProvider>
    );

    expect(triggerToast).toBeDefined();
    act(() => {
      triggerToast?.();
    });
    expect(screen.getByText('Ephemeral')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    expect(screen.queryByText('Ephemeral')).not.toBeInTheDocument();
  });
});
