import React from 'react';

export class ErrorBoundary extends React.Component {
  state = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return (
        <main className="error-page">
          <div className="error-card">
            <span className="eyebrow">BASHOCODE CHAT</span>
            <h1>Terjadi kesalahan pada tampilan.</h1>
            <button type="button" onClick={() => window.location.reload()}>Muat ulang</button>
          </div>
        </main>
      );
    }
    return this.props.children;
  }
}
