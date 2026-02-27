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

(function() {
    let miniSearch = null;
    let isIndexLoaded = false;
    let selectedIndex = -1; 
    
    function initSearch() {
        const searchModal = document.getElementById('docmd-search-modal');
        const searchInput = document.getElementById('docmd-search-input');
        const searchResults = document.getElementById('docmd-search-results');
        
        if (!searchModal) return;

        const rawRoot = window.DOCMD_ROOT || './';
        const ROOT_PATH = rawRoot.endsWith('/') ? rawRoot : rawRoot + '/';
        const emptyStateHtml = '<div class="search-initial">Type to start searching...</div>';

        // 1. Open/Close Logic
        function openSearch() {
            searchModal.style.display = 'flex';
            window.lastFocusedElement = document.activeElement;
            setTimeout(() => searchInput.focus(), 50);
            
            if (!searchInput.value.trim()) {
                searchResults.innerHTML = emptyStateHtml;
                selectedIndex = -1;
            }
            if (!isIndexLoaded) loadIndex();
        }

        function closeSearch() {
            searchModal.style.display = 'none';
            if (window.lastFocusedElement) window.lastFocusedElement.focus();
            selectedIndex = -1;
        }

        // --- Event Delegation for Triggers (Survives SPA) ---
        document.addEventListener('click', (e) => {
            if (e.target.closest('.docmd-search-trigger')) {
                e.preventDefault();
                openSearch();
            }
            if (e.target === searchModal || e.target.closest('.docmd-search-close')) {
                closeSearch();
            }
        });

        // 2. Keyboard Navigation
        document.addEventListener('keydown', (e) => {
            if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
                e.preventDefault();
                searchModal.style.display === 'flex' ? closeSearch() : openSearch();
            }
            
            if (searchModal.style.display === 'flex') {
                const items = searchResults.querySelectorAll('.search-result-item');
                if (e.key === 'Escape') { e.preventDefault(); closeSearch(); }
                else if (e.key === 'ArrowDown') { e.preventDefault(); if (items.length) { selectedIndex = (selectedIndex + 1) % items.length; updateSelection(items); } }
                else if (e.key === 'ArrowUp') { e.preventDefault(); if (items.length) { selectedIndex = (selectedIndex - 1 + items.length) % items.length; updateSelection(items); } }
                else if (e.key === 'Enter') {
                    e.preventDefault();
                    if (selectedIndex >= 0 && items[selectedIndex]) items[selectedIndex].click();
                    else if (items.length > 0) items[0].click();
                }
            }
        });

        function updateSelection(items) {
            items.forEach((item, idx) => {
                item.classList.toggle('selected', idx === selectedIndex);
                if (idx === selectedIndex) item.scrollIntoView({ block: 'nearest' });
            });
        }

        // 3. Index Loading
        async function loadIndex() {
            try {
                const indexUrl = `${ROOT_PATH}search-index.json`;
                const response = await fetch(indexUrl);
                if (response.headers.get("content-type")?.includes("text/html")) throw new Error("Invalid content type");
                if (!response.ok) throw new Error(response.status);

                const jsonString = await response.text();
                miniSearch = MiniSearch.loadJSON(jsonString, {
                    fields: ['title', 'headings', 'text'],
                    storeFields: ['title', 'id', 'text'],
                    searchOptions: { fuzzy: 0.2, prefix: true, boost: { title: 2, headings: 1.5 } }
                });
                isIndexLoaded = true;
                if (searchInput.value.trim()) searchInput.dispatchEvent(new Event('input'));
            } catch (e) {
                searchResults.innerHTML = '<div class="search-error">Failed to load search index.</div>';
            }
        }

        function getSnippet(text, query) {
            if (!text) return '';
            const terms = query.split(/\s+/).filter(t => t.length > 2);
            let bestIndex = -1;
            for (const term of terms) {
                const idx = text.toLowerCase().indexOf(term.toLowerCase());
                if (idx >= 0) { bestIndex = idx; break; }
            }
            let start = Math.max(0, bestIndex - 60);
            let end = Math.min(text.length, bestIndex + 60);
            let snippet = text.substring(start, end);
            if (start > 0) snippet = '...' + snippet;
            if (end < text.length) snippet += '...';
            
            const safeTerms = terms.map(t => t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|');
            if (safeTerms) {
                snippet = snippet.replace(new RegExp(`(${safeTerms})`, 'gi'), '<mark>$1</mark>');
            }
            return snippet;
        }

        searchInput.addEventListener('input', (e) => {
            const query = e.target.value.trim();
            selectedIndex = -1;
            if (!query) { searchResults.innerHTML = emptyStateHtml; return; }
            if (!isIndexLoaded) return;

            const results = miniSearch.search(query);
            if (results.length === 0) {
                searchResults.innerHTML = '<div class="search-no-results">No results found.</div>';
                return;
            }

            searchResults.innerHTML = results.slice(0, 10).map((result, index) => {
                const snippet = getSnippet(result.text, query);
                const linkHref = `${ROOT_PATH}${result.id}`;
                return `
                    <a href="${linkHref}" class="search-result-item" data-index="${index}">
                        <div class="search-result-title">${result.title}</div>
                        <div class="search-result-preview">${snippet}</div>
                    </a>`;
            }).join('');

            searchResults.querySelectorAll('.search-result-item').forEach((item, idx) => {
                item.addEventListener('mouseenter', () => { selectedIndex = idx; updateSelection(searchResults.querySelectorAll('.search-result-item')); });
            });
        });

        // Close search when clicking a link (Important for SPA!)
        searchResults.addEventListener('click', (e) => {
            if (e.target.closest('.search-result-item')) closeSearch();
        });
        
        window.closeDocmdSearch = closeSearch;
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initSearch);
    } else {
        initSearch();
    }
})();