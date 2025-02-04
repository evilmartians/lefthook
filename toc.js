// Populate the sidebar
//
// This is a script, and not included directly in the page, to control the total size of the book.
// The TOC contains an entry for each page, so if each page includes a copy of the TOC,
// the total size of the page becomes O(n**2).
class MDBookSidebarScrollbox extends HTMLElement {
    constructor() {
        super();
    }
    connectedCallback() {
        this.innerHTML = '<ol class="chapter"><li class="chapter-item affix "><a href="intro.html">Introduction</a></li><li class="chapter-item affix "><li class="part-title">User guide</li><li class="chapter-item "><a href="installation/index.html">Installation</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="installation/ruby.html">Ruby</a></li><li class="chapter-item "><a href="installation/node.html">Node.js</a></li><li class="chapter-item "><a href="installation/go.html">Go</a></li><li class="chapter-item "><a href="installation/python.html">Python</a></li><li class="chapter-item "><a href="installation/swift.html">Swift</a></li><li class="chapter-item "><a href="installation/scoop.html">Scoop</a></li><li class="chapter-item "><a href="installation/homebrew.html">Homebrew</a></li><li class="chapter-item "><a href="installation/winget.html">Winget</a></li><li class="chapter-item "><a href="installation/snap.html">Snap</a></li><li class="chapter-item "><a href="installation/deb.html">Debian-based distro</a></li><li class="chapter-item "><a href="installation/rpm.html">RPM-based distro</a></li><li class="chapter-item "><a href="installation/alpine.html">Alpine</a></li><li class="chapter-item "><a href="installation/arch.html">Arch Linux</a></li><li class="chapter-item "><a href="installation/manual.html">Manual</a></li></ol></li><li class="chapter-item "><a href="usage/index.html">Usage</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="usage/commands.html">Commands</a></li><li class="chapter-item "><a href="usage/env.html">ENV variables</a></li><li class="chapter-item "><a href="usage/tips.html">Tips</a></li></ol></li><li class="chapter-item "><li class="part-title">Reference guide</li><li class="chapter-item "><a href="configuration/index.html">Configuration</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="configuration/assert_lefthook_installed.html">assert_lefthook_installed</a></li><li class="chapter-item "><a href="configuration/colors.html">colors</a></li><li class="chapter-item "><a href="configuration/extends.html">extends</a></li><li class="chapter-item "><a href="configuration/lefthook.html">lefthook</a></li><li class="chapter-item "><a href="configuration/min_version.html">min_version</a></li><li class="chapter-item "><a href="configuration/no_tty.html">no_tty</a></li><li class="chapter-item "><a href="configuration/output.html">output</a></li><li class="chapter-item "><a href="configuration/rc.html">rc</a></li><li class="chapter-item "><a href="configuration/remotes.html">remotes</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="configuration/git_url.html">git_url</a></li><li class="chapter-item "><a href="configuration/ref.html">ref</a></li><li class="chapter-item "><a href="configuration/refetch.html">refetch</a></li><li class="chapter-item "><a href="configuration/refetch_frequency.html">refetch_frequency</a></li><li class="chapter-item "><a href="configuration/configs.html">configs</a></li></ol></li><li class="chapter-item "><a href="configuration/skip_output.html">skip_output</a></li><li class="chapter-item "><a href="configuration/source_dir.html">source_dir</a></li><li class="chapter-item "><a href="configuration/source_dir_local.html">source_dir_local</a></li><li class="chapter-item "><a href="configuration/skip_lfs.html">skip_lfs</a></li><li class="chapter-item "><a href="configuration/templates.html">templates</a></li><li class="chapter-item "><a href="configuration/Hook.html">{Git hook name}</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="configuration/files-global.html">files</a></li><li class="chapter-item "><a href="configuration/parallel.html">parallel</a></li><li class="chapter-item "><a href="configuration/piped.html">piped</a></li><li class="chapter-item "><a href="configuration/follow.html">follow</a></li><li class="chapter-item "><a href="configuration/exclude_tags.html">exclude_tags</a></li><li class="chapter-item "><a href="configuration/skip.html">skip</a></li><li class="chapter-item "><a href="configuration/only.html">only</a></li><li class="chapter-item "><a href="configuration/jobs.html">jobs</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="configuration/name.html">name</a></li><li class="chapter-item "><a href="configuration/run.html">run</a></li><li class="chapter-item "><a href="configuration/script.html">script</a></li><li class="chapter-item "><a href="configuration/runner.html">runner</a></li><li class="chapter-item "><a href="configuration/group.html">group</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="configuration/parallel.html">parallel</a></li><li class="chapter-item "><a href="configuration/piped.html">piped</a></li><li class="chapter-item "><a href="configuration/jobs.html">jobs</a></li></ol></li><li class="chapter-item "><a href="configuration/skip.html">skip</a></li><li class="chapter-item "><a href="configuration/only.html">only</a></li><li class="chapter-item "><a href="configuration/tags.html">tags</a></li><li class="chapter-item "><a href="configuration/glob.html">glob</a></li><li class="chapter-item "><a href="configuration/files.html">files</a></li><li class="chapter-item "><a href="configuration/file_types.html">file_types</a></li><li class="chapter-item "><a href="configuration/env.html">env</a></li><li class="chapter-item "><a href="configuration/root.html">root</a></li><li class="chapter-item "><a href="configuration/exclude.html">exclude</a></li><li class="chapter-item "><a href="configuration/fail_text.html">fail_text</a></li><li class="chapter-item "><a href="configuration/stage_fixed.html">stage_fixed</a></li><li class="chapter-item "><a href="configuration/interactive.html">interactive</a></li><li class="chapter-item "><a href="configuration/use_stdin.html">use_stdin</a></li></ol></li><li class="chapter-item "><a href="configuration/Commands.html">commands</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="configuration/run.html">run</a></li><li class="chapter-item "><a href="configuration/skip.html">skip</a></li><li class="chapter-item "><a href="configuration/only.html">only</a></li><li class="chapter-item "><a href="configuration/tags.html">tags</a></li><li class="chapter-item "><a href="configuration/glob.html">glob</a></li><li class="chapter-item "><a href="configuration/files.html">files</a></li><li class="chapter-item "><a href="configuration/file_types.html">file_types</a></li><li class="chapter-item "><a href="configuration/env.html">env</a></li><li class="chapter-item "><a href="configuration/root.html">root</a></li><li class="chapter-item "><a href="configuration/exclude.html">exclude</a></li><li class="chapter-item "><a href="configuration/fail_text.html">fail_text</a></li><li class="chapter-item "><a href="configuration/stage_fixed.html">stage_fixed</a></li><li class="chapter-item "><a href="configuration/interactive.html">interactive</a></li><li class="chapter-item "><a href="configuration/use_stdin.html">use_stdin</a></li><li class="chapter-item "><a href="configuration/priority.html">priority</a></li></ol></li><li class="chapter-item "><a href="configuration/Scripts.html">scripts</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="configuration/runner.html">runner</a></li><li class="chapter-item "><a href="configuration/skip.html">skip</a></li><li class="chapter-item "><a href="configuration/only.html">only</a></li><li class="chapter-item "><a href="configuration/tags.html">tags</a></li><li class="chapter-item "><a href="configuration/env.html">env</a></li><li class="chapter-item "><a href="configuration/fail_text.html">fail_text</a></li><li class="chapter-item "><a href="configuration/stage_fixed.html">stage_fixed</a></li><li class="chapter-item "><a href="configuration/interactive.html">interactive</a></li><li class="chapter-item "><a href="configuration/use_stdin.html">use_stdin</a></li><li class="chapter-item "><a href="configuration/priority.html">priority</a></li></ol></li></ol></li></ol></li><li class="chapter-item "><a href="examples/index.html">Examples</a><a class="toggle"><div>❱</div></a></li><li><ol class="section"><li class="chapter-item "><a href="examples/lefthook-local.html">lefthook-local.yml</a></li><li class="chapter-item "><a href="examples/stage_fixed.html">Auto stage changed files</a></li><li class="chapter-item "><a href="examples/filters.html">Filter files</a></li><li class="chapter-item "><a href="examples/skip.html">Skip or run on condition</a></li><li class="chapter-item "><a href="examples/remotes.html">Use remote config</a></li><li class="chapter-item "><a href="examples/commitlint.html">Use commitlint</a></li></ol></li><li class="chapter-item "><li class="spacer"></li><li class="chapter-item affix "><a href="misc/contributors.html">Contributors</a></li></ol>';
        // Set the current, active page, and reveal it if it's hidden
        let current_page = document.location.href.toString().split("#")[0];
        if (current_page.endsWith("/")) {
            current_page += "index.html";
        }
        var links = Array.prototype.slice.call(this.querySelectorAll("a"));
        var l = links.length;
        for (var i = 0; i < l; ++i) {
            var link = links[i];
            var href = link.getAttribute("href");
            if (href && !href.startsWith("#") && !/^(?:[a-z+]+:)?\/\//.test(href)) {
                link.href = path_to_root + href;
            }
            // The "index" page is supposed to alias the first chapter in the book.
            if (link.href === current_page || (i === 0 && path_to_root === "" && current_page.endsWith("/index.html"))) {
                link.classList.add("active");
                var parent = link.parentElement;
                if (parent && parent.classList.contains("chapter-item")) {
                    parent.classList.add("expanded");
                }
                while (parent) {
                    if (parent.tagName === "LI" && parent.previousElementSibling) {
                        if (parent.previousElementSibling.classList.contains("chapter-item")) {
                            parent.previousElementSibling.classList.add("expanded");
                        }
                    }
                    parent = parent.parentElement;
                }
            }
        }
        // Track and set sidebar scroll position
        this.addEventListener('click', function(e) {
            if (e.target.tagName === 'A') {
                sessionStorage.setItem('sidebar-scroll', this.scrollTop);
            }
        }, { passive: true });
        var sidebarScrollTop = sessionStorage.getItem('sidebar-scroll');
        sessionStorage.removeItem('sidebar-scroll');
        if (sidebarScrollTop) {
            // preserve sidebar scroll position when navigating via links within sidebar
            this.scrollTop = sidebarScrollTop;
        } else {
            // scroll sidebar to current active section when navigating via "next/previous chapter" buttons
            var activeSection = document.querySelector('#sidebar .active');
            if (activeSection) {
                activeSection.scrollIntoView({ block: 'center' });
            }
        }
        // Toggle buttons
        var sidebarAnchorToggles = document.querySelectorAll('#sidebar a.toggle');
        function toggleSection(ev) {
            ev.currentTarget.parentElement.classList.toggle('expanded');
        }
        Array.from(sidebarAnchorToggles).forEach(function (el) {
            el.addEventListener('click', toggleSection);
        });
    }
}
window.customElements.define("mdbook-sidebar-scrollbox", MDBookSidebarScrollbox);
