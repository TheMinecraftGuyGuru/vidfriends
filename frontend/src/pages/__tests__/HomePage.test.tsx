import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import { ToastProvider } from '../../components/ToastProvider';
import { AppStateProvider } from '../../state/AppStateProvider';
import { HomePage } from '../HomePage';

describe('HomePage', () => {
  it('renders the welcome headline', () => {
    render(
      <ToastProvider>
        <AppStateProvider>
          <MemoryRouter>
            <HomePage />
          </MemoryRouter>
        </AppStateProvider>
      </ToastProvider>
    );

    expect(screen.getByText(/welcome to vidfriends/i)).toBeInTheDocument();
  });
});
