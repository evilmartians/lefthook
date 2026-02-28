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
  let counter = 0;

  function getTheme() {
    return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'default';
  }

  async function renderAll() {
    mermaid.initialize({ startOnLoad: false, theme: getTheme(), securityLevel: 'loose' });

    const elements = document.querySelectorAll('.mermaid:not([data-processed="true"])');
    
    for (const el of elements) {
      if (!el.dataset.original) el.dataset.original = el.textContent;
      const code = el.dataset.original;

      // Skip elements that are strictly display:none (Wait for tab click)
      if (el.offsetParent === null) continue;

      try {
        const id = `mermaid-svg-${counter++}`;
        // Generate SVG string in memory (prevents D3 bounding box crashes)
        const { svg } = await mermaid.render(id, code);
        el.innerHTML = svg;
        el.setAttribute('data-processed', 'true');
      } catch (e) {
        el.setAttribute('data-processed', 'error');
      }
    }
  }

  // 1. Initial Load
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', renderAll);
  } else {
    renderAll();
  }

  // 2. SPA Navigation Load
  document.addEventListener('docmd:page-mounted', renderAll);

  // 3. Render when a hidden Tab or Collapsible is opened
  document.addEventListener('click', (e) => {
    if (e.target.closest('.docmd-tabs-nav-item, .collapsible-summary')) {
      setTimeout(renderAll, 50); // Small delay to let CSS apply display:block
    }
  });

  // 4. Theme Toggle
  const themeObserver = new MutationObserver((mutations) => {
    for (const m of mutations) {
      if (m.attributeName === 'data-theme') {
        document.querySelectorAll('.mermaid').forEach(el => el.removeAttribute('data-processed'));
        renderAll();
      }
    }
  });
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter:['data-theme'] });

})();