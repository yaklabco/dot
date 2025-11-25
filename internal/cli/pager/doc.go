// Package pager provides terminal pagination for long content.
//
// The pager displays content page-by-page when it exceeds the terminal height,
// allowing users to navigate through long output without losing context.
//
// Features:
//   - Automatic terminal size detection
//   - Page-by-page navigation
//   - Line count awareness
//   - Graceful fallback for non-interactive terminals
package pager
