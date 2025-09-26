import { Suspense } from 'react';

import { ErrorBoundary } from './components/ErrorBoundary';
import { AppRouter } from './routes/AppRouter';

function App() {
  return (
    <ErrorBoundary>
      <Suspense fallback={<div className="app-loading">Loading VidFriends...</div>}>
        <AppRouter />
      </Suspense>
    </ErrorBoundary>
  );
}

export default App;
