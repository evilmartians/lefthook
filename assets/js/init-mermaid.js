/**
 * --------------------------------------------------------------------
 * docmd : the minimalist, zero-config documentation generator.
 *
 * @package     @docmd/core (and ecosystem)
 * @website     https://docmd.io
 * @repository  https://github.com/docmd-io/docmd
 * @license     MIT
 * @copyright   Copyright (c) 2025 docmd.io
 *
 * [docmd-source] - Please do not remove this header.
 * --------------------------------------------------------------------
 */

import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs';

(async function () {
  'use strict';

  function getTheme() {
    return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'default';
  }

  function backupOriginals() {
    document.querySelectorAll('.mermaid').forEach(el => {
      // textContent automatically decodes the HTML escaped entities back to > and <
      if (!el.dataset.original) el.dataset.original = el.textContent;
    });
  }

  // 1. Lazy Renderer: Only renders elements that are currently visible on screen
  const renderObserver = new IntersectionObserver((entries) => {
    const visibleNodes = entries
      .filter(entry => entry.isIntersecting && entry.target.offsetWidth > 0)
      .map(entry => entry.target);
      
    if (visibleNodes.length > 0) {
      // Stop observing these specific nodes since we are rendering them now
      visibleNodes.forEach(node => renderObserver.unobserve(node));
      
      mermaid.initialize({ startOnLoad: false, theme: getTheme(), securityLevel: 'loose' });
      mermaid.run({ nodes: visibleNodes }).catch(e => console.warn('Mermaid render skipped for hidden element'));
    }
  });

  function observeAll() {
    backupOriginals();
    document.querySelectorAll('.mermaid:not([data-processed="true"])').forEach(el => {
      // Check if it's already visible. If so, IntersectionObserver will catch it immediately.
      renderObserver.observe(el);
    });
  }

  // 2. Theme Toggle Handler
  const themeObserver = new MutationObserver((mutations) => {
    for (const m of mutations) {
      if (m.attributeName === 'data-theme') {
        document.querySelectorAll('.mermaid').forEach(el => {
          el.removeAttribute('data-processed');
          el.textContent = el.dataset.original;
          renderObserver.observe(el);
        });
      }
    }
  });
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });

  // 3. Bootstrapping
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', observeAll);
  } else {
    observeAll();
  }

  // 4. Hook into SPA Router
  document.addEventListener('docmd:page-mounted', observeAll);
  
  // 5. Hook into Tabs/Collapsible clicks to trigger instant render when unhidden
  document.addEventListener('click', (e) => {
    if (e.target.closest('.docmd-tabs-nav-item, .collapsible-summary')) {
        // Wait 50ms for the browser to apply display: block to the tab pane
        setTimeout(() => {
            const nodes = Array.from(document.querySelectorAll('.mermaid:not([data-processed="true"])'))
                               .filter(n => n.offsetWidth > 0);
            if (nodes.length > 0) {
                nodes.forEach(n => renderObserver.unobserve(n));
                mermaid.initialize({ startOnLoad: false, theme: getTheme(), securityLevel: 'loose' });
                mermaid.run({ nodes }).catch(err => {});
            }
        }, 50);
    }
  });
})();