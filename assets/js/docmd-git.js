/**
 * --------------------------------------------------------------------
 * docmd : the zero-config documentation engine.
 *
 * @package     @docmd/core (and ecosystem)
 * @website     https://docmd.io
 * @repository  https://github.com/docmd-io/docmd
 * @license     MIT
 * @copyright   Copyright (c) 2025-present docmd.io
 *
 * [docmd-source] - Please do not remove this header.
 * --------------------------------------------------------------------
 */

(function () {
  const i18n = window.__git_i18n || {};

  /**
   * Format a Unix ms timestamp into a human-readable relative or absolute string.
   */
  function formatTime(ts) {
    const config = window.__git_config || {};
    const format = config.dateFormat || 'relative';
    const dateObj = new Date(ts);

    if (format === 'iso') {
      return dateObj.toISOString();
    }
    
    if (format === 'locale-aware') {
      return dateObj.toLocaleString(document.documentElement.lang || 'en');
    }

    const diff = Date.now() - ts;
    const s = Math.floor(diff / 1000);
    const m = Math.floor(s / 60);
    const h = Math.floor(m / 60);
    const d = Math.floor(h / 24);

    if (d < 1) {
      if (h >= 1) return (i18n.hoursAgo || '{n}h ago').replace('{n}', h);
      if (m >= 1) return (i18n.minutesAgo || '{n}m ago').replace('{n}', m);
      return i18n.justNow || 'just now';
    }
    if (d < 7) return (i18n.daysAgo || '{n}d ago').replace('{n}', d);

    return dateObj.toLocaleDateString(
      document.documentElement.lang || 'en',
      { year: 'numeric', month: 'short', day: 'numeric' }
    );
  }

  /**
   * Hydrate all git [data-timestamp] elements with formatted relative time.
   * Scoped to .git-last-updated and .git-commit-list to avoid collisions.
   */
  function hydrateTimes() {
    document.querySelectorAll('.git-last-updated [data-timestamp], .git-commit-list [data-timestamp]').forEach(function (el) {
      const ts = parseInt(el.getAttribute('data-timestamp'), 10);
      if (!ts || isNaN(ts)) return;
      el.textContent = formatTime(ts);
    });
  }

  /**
   * Wire up keyboard accessibility on .git-last-updated elements.
   * Tooltip visibility is handled by CSS :hover/:focus-within.
   * JS only adds tabindex and aria attributes for keyboard users.
   */
  function wireTooltips() {
    document.querySelectorAll('.git-last-updated').forEach(function (wrapper) {
      const tooltip = wrapper.querySelector('.git-commit-tooltip');
      if (!tooltip) return;

      wrapper.classList.add('has-history');
      wrapper.setAttribute('tabindex', '0');
      wrapper.setAttribute('role', 'button');
      wrapper.setAttribute('aria-label', i18n.viewCommitHistory || 'View commit history');
    });
  }

  function init() {
    hydrateTimes();
    wireTooltips();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  // Re-run after SPA navigation — the SPA router swaps .page-footer-actions
  // innerHTML on every page change, so we must re-hydrate timestamps and
  // re-wire tooltip accessibility on the freshly injected elements.
  document.addEventListener('docmd:page-mounted', init);
})();