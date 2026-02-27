// docmd.config.js
module.exports = {
  // --- Core Metadata ---
  siteTitle: 'Lefthook',
  siteUrl: '', // e.g. https://mysite.com (Critical for SEO/Sitemap)

  // --- Branding ---
  logo: {
    light: '/assets/lefthook.png',
    dark: '/assets/lefthook.png',
    alt: 'Logo',
    href: './',
  },
  favicon: '/assets/lefthook.png',

  // --- Source & Output ---
  srcDir: 'docs',
  outputDir: 'site',

  // --- Theme & Layout ---
  theme: {
    name: 'default',            // Options: 'default', 'sky', 'ruby', 'retro'
    defaultMode: 'system',  // 'light', 'dark', or 'system'
    enableModeToggle: true, // Show mode toggle button
    positionMode: 'top',    // 'top' or 'bottom'
    codeHighlight: true,    // Enable Highlight.js
    customCss: ['assets/css/lefthook.css'],          // e.g. ['assets/css/custom.css']
  },

  // --- Features ---
  search: true,           // Built-in offline search
  minify: true,           // Minify HTML/CSS/JS in build
  autoTitleFromH1: true,  // Auto-generate page title from first H1
  copyCode: true,         // Show "copy" button on code blocks
  pageNavigation: false,   // Prev/Next buttons at bottom

  // --- Navigation (Sidebar) ---
  navigation: [
    {
      title: "Installation",
      icon: 'rocket',
      path: "/install",
      collapsible: true,
      children: [
        { title: "Ruby", path: "/installation/ruby" },
        { title: "Node.js", path: "/installation/node" },
        { title: "Swift", path: "/installation/swift" },
        { title: "Go", path: "/installation/go" },
        { title: "Python", path: "/installation/python" },
        { title: "Scoop", path: "/installation/scoop" },
        { title: "Homebrew", path: "/installation/homebrew" },
        { title: "Winget", path: "/installation/winget" },
        { title: "Snap", path: "/installation/snap" },
        { title: "Debian-based distro", path: "/installation/deb" },
        { title: "RPM-based distro", path: "/installation/rpm" },
        { title: "Alpine", path: "/installation/alpine" },
        { title: "Arch Linux", path: "/installation/arch" },
        { title: "Mise", path: "/installation/mise" },
        { title: "Manual", path: "/installation/manual" },
      ],
    },
    {
      title: "Configuration",
      path: "/configuration",
      collapsible: true,
      icon: 'settings',
      children: [
        { title: "assert_lefthook_installed", path: "/configuration/assert_lefthook_installed" },
        { title: "colors", path: "/configuration/colors" },
        { title: "extends", path: "/configuration/extends" },
        { title: "install_non_git_hooks", path: "/configuration/install_non_git_hooks" },
        { title: "lefthook", path: "/configuration/lefthook" },
        { title: "min_version", path: "/configuration/min_version" },
        { title: "no_auto_install", path: "/configuration/no_auto_install" },
        { title: "no_tty", path: "/configuration/no_tty" },
        { title: "output", path: "/configuration/output" },
        { title: "rc", path: "/configuration/rc" },
        { title: "remotes", path: "/configuration/remotes",
          children: [
            { title: "git_url", path: "/configuration/git_url" },
            { title: "ref", path: "/configuration/ref" },
            { title: "refetch", path: "/configuration/refetch" },
            { title: "refetch_frequency", path: "/configuration/refetch_frequency" },
            { title: "configs", path: "/configuration/configs" },
          ]
        },
        { title: "source_dir", path: "/configuration/source_dir" },
        { title: "source_dir_local", path: "/configuration/source_dir_local" },
        { title: "skip_lfs", path: "/configuration/skip_lfs" },
        { title: "glob_matcher", path: "/configuration/glob_matcher" },
        { title: "templates", path: "/configuration/templates" },
        { title: "Hook", path: "/configuration/Hook",
          children: [
            { title: "files", path: "/configuration/files-global" },
            { title: "parallel", path: "/configuration/parallel" },
            { title: "piped", path: "/configuration/piped" },
            { title: "follow", path: "/configuration/follow" },
            { title: "fail_on_changes", path: "/configuration/fail_on_changes" },
            { title: "fail_on_changes_diff", path: "/configuration/fail_on_changes_diff" },
            { title: "exclude_tags", path: "/configuration/exclude_tags" },
            { title: "exclude", path: "/configuration/exclude" },
            { title: "skip", path: "/configuration/skip" },
            { title: "only", path: "/configuration/only" },
            { title: "jobs", path: "/configuration/jobs" ,
              children: [
                { title: "name", path: "/configuration/name" },
                { title: "run", path: "/configuration/run" },
                { title: "script", path: "/configuration/script" },
                { title: "runner", path: "/configuration/runner" },
                { title: "args", path: "/configuration/args" },
                { title: "group", collapsible: true, path: "/configuration/group",
                  children: [
                    { title: "parallel", path: "/configuration/parallel" },
                    { title: "piped", path: "/configuration/piped" },
                    { title: "jobs", path: "/configuration/jobs" },
                  ],
                },
                { title: "skip", path: "/configuration/skip" },
                { title: "only", path: "/configuration/only" },
                { title: "tags", path: "/configuration/tags" },
                { title: "glob", path: "/configuration/glob" },
                { title: "files", path: "/configuration/files" },
                { title: "file_types", path: "/configuration/file_types" },
                { title: "env", path: "/configuration/env" },
                { title: "root", path: "/configuration/root" },
                { title: "exclude", path: "/configuration/exclude" },
                { title: "fail_text", path: "/configuration/fail_text" },
                { title: "stage_fixed", path: "/configuration/stage_fixed" },
                { title: "interactive", path: "/configuration/interactive" },
                { title: "use_stdin", path: "/configuration/use_stdin" },
              ],
            },
            { title: "commands", path: "/configuration/Commands",
              children: [
                { title: "run", path: "/configuration/run" },
                { title: "skip", path: "/configuration/skip" },
                { title: "only", path: "/configuration/only" },
                { title: "tags", path: "/configuration/tags" },
                { title: "glob", path: "/configuration/glob" },
                { title: "files", path: "/configuration/files" },
                { title: "file_types", path: "/configuration/file_types" },
                { title: "env", path: "/configuration/env" },
                { title: "root", path: "/configuration/root" },
                { title: "exclude", path: "/configuration/exclude" },
                { title: "fail_text", path: "/configuration/fail_text" },
                { title: "stage_fixed", path: "/configuration/stage_fixed" },
                { title: "interactive", path: "/configuration/interactive" },
                { title: "use_stdin", path: "/configuration/use_stdin" },
                { title: "priority", path: "/configuration/priority" },
              ],
            },
            { title: "scripts", path: "/configuration/Scripts",
              children: [
                { title: "runner", path: "/configuration/runner" },
                { title: "args", path: "/configuration/args" },
                { title: "skip", path: "/configuration/skip" },
                { title: "only", path: "/configuration/only" },
                { title: "tags", path: "/configuration/tags" },
                { title: "env", path: "/configuration/env" },
                { title: "fail_text", path: "/configuration/fail_text" },
                { title: "stage_fixed", path: "/configuration/stage_fixed" },
                { title: "interactive", path: "/configuration/interactive" },
                { title: "use_stdin", path: "/configuration/use_stdin" },
                { title: "priority", path: "/configuration/priority" },
              ],
            },
          ],
        },
      ],
    },
    { title: "CLI", collapsible: true, icon: "terminal",
      children: [
            { title: "lefthook install", icon: "chevron-right", path: "/usage/commands/install" },
            { title: "lefthook uninstall", icon: "chevron-right", path: "/usage/commands/uninstall" },
            { title: "lefthook run", icon: "chevron-right", path: "/usage/commands/run" },
            { title: "lefthook add", icon: "chevron-right", path: "/usage/commands/add" },
            { title: "lefthook validate", icon: "chevron-right", path: "/usage/commands/validate" },
            { title: "lefthook dump", icon: "chevron-right", path: "/usage/commands/dump" },
            { title: "lefthook check-install", icon: "chevron-right", path: "/usage/commands/check-install" },
            { title: "lefthook self-update", icon: "chevron-right", path: "/usage/commands/self-update" },
        { title: "ENV variables", collapsible: true, icon: "dollar-sign",
          children: [
            { title: "LEFTHOOK", path: "/usage/envs/LEFTHOOK" },
            { title: "LEFTHOOK_VERBOSE", path: "/usage/envs/LEFTHOOK_VERBOSE" },
            { title: "LEFTHOOK_OUTPUT", path: "/usage/envs/LEFTHOOK_OUTPUT" },
            { title: "LEFTHOOK_CONFIG", path: "/usage/envs/LEFTHOOK_CONFIG" },
            { title: "LEFTHOOK_EXCLUDE", path: "/usage/envs/LEFTHOOK_EXCLUDE" },
            { title: "CLICOLOR_FORCE", path: "/usage/envs/CLICOLOR_FORCE" },
            { title: "NO_COLOR", path: "/usage/envs/NO_COLOR" },
            { title: "CI", path: "/usage/envs/CI" },
          ],
        },
      ],
    },
    { title: "Examples", collapsible: true, icon: "file-code",
      children: [
        { title: "Using local only config", path: "/examples/lefthook-local" },
        { title: "Wrap commands locally", path: "/examples/wrap-commands" },
        { title: "Auto add linter fixes to commit", path: "/examples/stage_fixed" },
        { title: "Filter files", path: "/examples/filters" },
        { title: "Skip or run on condition", path: "/examples/skip" },
        { title: "Remote configs", path: "/examples/remotes" },
        { title: "With commitlint", path: "/examples/commitlint" },
      ],
    },
    { title: "Contributors", path: "/misc/contributors", icon: "users-round" },
    { title: 'GitHub', path: 'https://github.com/evilmartians/lefthook', icon: 'github', external: true },
  ],

  // --- Plugins ---
  plugins: {
    seo: {
      defaultDescription: 'Lefthook documentation.',
      openGraph: {
        defaultImage: '',   // e.g. 'assets/images/og-image.png'
      },
      // twitter: {
      //   cardType: 'summary_large_image',
      // }
    },
    analytics: {
      // googleV4: {
      //   measurementId: 'G-X9WTDL262N' // Replace with your GA Measurement ID
      // }
    },
    sitemap: {
      defaultChangefreq: 'weekly',  // e.g. 'daily', 'weekly', 'monthly'
      defaultPriority: 0.8          // Priority between 0.0 and 1.0
    },
    search: {},
    mermaid: {},
    llms: {}
  },

  // --- Footer ---
  footer: '© ' + new Date().getFullYear() + ' Lefthook.',

  // --- Edit Link ---
  editLink: {
    enabled: false,
    baseUrl: 'https://github.com/evilmartians/lefthook/edit/main/docs',
    text: 'Edit this page'
  }
};
