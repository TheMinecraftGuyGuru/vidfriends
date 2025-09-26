import { Component, ErrorInfo, ReactNode } from 'react';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('VidFriends encountered an unrecoverable error.', error, info);
  }

  handleReload = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="app-error-boundary" role="alert">
          <div className="app-error-boundary__content">
            <h1>Something went wrong</h1>
            <p>We hit an unexpected issue loading VidFriends. Reload to try again.</p>
            <pre>{this.state.error?.message}</pre>
            <button type="button" onClick={this.handleReload}>
              Reload VidFriends
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
