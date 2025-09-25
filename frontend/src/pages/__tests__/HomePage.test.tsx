import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import { AppStateProvider } from '../../state/AppStateProvider';
import { HomePage } from '../HomePage';

describe('HomePage', () => {
  it('renders the welcome headline', () => {
    render(
      <AppStateProvider>
        <MemoryRouter>
          <HomePage />
        </MemoryRouter>
      </AppStateProvider>
    );

    expect(screen.getByText(/welcome to vidfriends/i)).toBeInTheDocument();
  });
});
