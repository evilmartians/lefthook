# Change log

## master (unreleased)

## 1.5.0 (2023-09-21)

- chore: output enhancements ([#549](https://github.com/evilmartians/lefthook/pull/549)) by @mrexox
- feat: add interrupt (Ctrl-C) handling ([#550](https://github.com/evilmartians/lefthook/pull/550)) by @mrcljx

## 1.4.11 (2023-09-13)

- docs: update docs and readme with tl;dr instructions ([#548](https://github.com/evilmartians/lefthook/pull/548)) by @mrexox
- fix: add use_stdin option for just reading from stdin ([#547](https://github.com/evilmartians/lefthook/pull/547)) by @mrexox
- chore: refactor commands passing ([#546](https://github.com/evilmartians/lefthook/pull/546)) by @mrexox
- fix: fail on non existing hook name ([#545](https://github.com/evilmartians/lefthook/pull/545)) by @mrexox

## 1.4.10 (2023-09-04)

- fix: split command with file templates into chunks ([#541](https://github.com/evilmartians/lefthook/pull/541)) by @mrexox
- chore: add git-cliff config for easier changelog generation by @mrexox
- fix: allow empty staged files diffs ([#543](https://github.com/evilmartians/lefthook/pull/543)) by @mrexox

## 1.4.9 (2023-08-15)

- chore: fix linter issues ([#537](https://github.com/evilmartians/lefthook/pull/537)) by @mrexox
- feat: add files, all-files, and commands flags ([#534](https://github.com/evilmartians/lefthook/pull/534)) by @nihalgonsalves
- chore: bump go to 1.21 ([#536](https://github.com/evilmartians/lefthook/pull/536)) by @nihalgonsalves

## 1.4.8 (2023-07-31)

- feat: add assert_lefthook_installed option ([#533](https://github.com/evilmartians/lefthook/pull/533)) by @mrexox
- chore: add *Add docs* to PR template ([#532](https://github.com/evilmartians/lefthook/pull/532)) by @technicalpickles
- feat: add support for skipping empty summaries ([#531](https://github.com/evilmartians/lefthook/pull/531)) by @technicalpickles

## 1.4.7 (2023-07-24)

- docs: add scoop installation method ([#527](https://github.com/evilmartians/lefthook/pull/527)) by @sitiom
- fix: correct merging of extends from remote config ([#529](https://github.com/evilmartians/lefthook/pull/529)) by @mrexox
- ci: add Winget Releaser action ([#526](https://github.com/evilmartians/lefthook/pull/526)) by @sitiom
- chore: improve correctness of load_test.go ([#525](https://github.com/evilmartians/lefthook/pull/525)) by @hyperupcall

## 1.4.6 (2023-07-18)

- fix: do not print extraneous newlines when executionInfo output is hidden ([#519](https://github.com/evilmartians/lefthook/pull/519)) by @hyperupcall
- fix: uninstall all possible formats ([#523](https://github.com/evilmartians/lefthook/pull/523)) by @mrexox
- fix: LEFTHOOK_VERBOSE properly overrides --verbose flag ([#521](https://github.com/evilmartians/lefthook/pull/521)) by @hyperupcall
- feat: support .lefthook.yml and .lefthook-local.yml ([#520](https://github.com/evilmartians/lefthook/pull/520)) by @hyperupcall

## 1.4.5 (2023-07-12)

- docs: improve documentation and examples ([#517](https://github.com/evilmartians/lefthook/pull/517)) by @hyperupcall
- fix: improve hook template ([#516](https://github.com/evilmartians/lefthook/pull/516)) by @hyperupcall

## 1.4.4 (2023-07-10)

- fix: don't render bold ANSI sequence when colors are disabled ([#515](https://github.com/evilmartians/lefthook/pull/515)) by @adam12
- deps: July 2023 ([#514](https://github.com/evilmartians/lefthook/pull/514)) by @mrexox

## 1.4.3 (2023-06-19)

- fix: auto stage non-standard files ([#506](https://github.com/evilmartians/lefthook/pull/506)) by @mrexox

## 1.4.2 (2023-06-13)

- deps: June 2023 ([#499](https://github.com/evilmartians/lefthook/pull/499))
- feat: support toml dumpint ([#490](https://github.com/evilmartians/lefthook/pull/490)) by @mrexox
- feat: support json configs ([#489](https://github.com/evilmartians/lefthook/pull/489)) by @mrexox

## 1.4.1 (2023-05-22)

- fix: add win32 binary to artifacts (by @mrexox)
- feat: allow dumping with JSON ([#485](https://github.com/evilmartians/lefthook/pull/485) by @mrexox
- feat: add skip execution_info option ([#484](https://github.com/evilmartians/lefthook/pull/484)) by @mrexox
- deps: from 05.2023 ([#487](https://github.com/evilmartians/lefthook/pull/487)) by @mrexox

## 1.4.0 (2023-05-18)

- feat: add adaptive colors ([#482](https://github.com/evilmartians/lefthook/pull/482)) by @mrexox
- fix: skip output for interactive commands if configured ([#483](https://github.com/evilmartians/lefthook/pull/483)) by @mrexox
- feat: add dump command ([#481](https://github.com/evilmartians/lefthook/pull/481)) by @mrexox

## 1.3.13 (2023-05-11)

- feat: add only option ([#478](https://github.com/evilmartians/lefthook/pull/478)) by @mrexox

## 1.3.12 (2023-04-28)

- fix: allow skipping execution_out with interactive mode ([#476](https://github.com/evilmartians/lefthook/pull/476)) by @mrexox

## 1.3.11 (2023-04-27)

- feat: add execution_out to skip output settings ([#475](https://github.com/evilmartians/lefthook/pull/475)) by @mrexox
- chore: add debug logs when hook is skipped ([#474](https://github.com/evilmartians/lefthook/pull/474)) by @mrexox

## 1.3.10

- feat: don't show when commands are skipped because of no matched files ([#468](https://github.com/evilmartians/lefthook/pull/468)) by @technicalpickles

## 1.3.9 (2023-04-04)

- feat: allow extra hooks in local config ([#462](https://github.com/evilmartians/lefthook/pull/462)) by @fabn
- feat: pass numeric placeholders to files command ([#461](https://github.com/evilmartians/lefthook/pull/461)) by @fabn

## 1.3.8 (2023-03-23)

- fix: make hook template compatible with shells without source command ([#458](https://github.com/evilmartians/lefthook/pull/458)) by @mdesantis

## 1.3.7 (2023-03-20)

- fix: allow globs in skip option ([#457](https://github.com/evilmartians/lefthook/pull/457)) by @mrexox
- deps: dependencies update (March 2023) ([#455](https://github.com/evilmartians/lefthook/pull/455)) by @mrexox
- fix: don't fail on missing config file ([#450](https://github.com/evilmartians/lefthook/pull/450)) by @mrexox

## 1.3.6 (2023-03-16)

- fix: stage fixed when root specified ([#449](https://github.com/evilmartians/lefthook/pull/449)) by @mrexox
- feat: implitic skip on missing files for pre-commit and pre-push hooks ([#448](https://github.com/evilmartians/lefthook/pull/448)) by @mrexox

## 1.3.5 (2023-03-15)

- feat: add stage_fixed option ([#445](https://github.com/evilmartians/lefthook/pull/445)) by @mrexox

## 1.3.4 (2023-03-13)

- fix: don't extra extend config if lefthook-local.yml is missing ([#444](https://github.com/evilmartians/lefthook/pull/444)) by @mrexox

## 1.3.3 (2023-03-01)

- fix: restore release assets name ([#437](https://github.com/evilmartians/lefthook/pull/437)) by @watarukura

## 1.3.2 (2023-02-27)

- fix: Allow sh syntax in files ([#435](https://github.com/evilmartians/lefthook/pull/435)) by @mrexox

## 1.3.1 (2023-02-27)

- fix: Force creation of git hooks folder ([#434](https://github.com/evilmartians/lefthook/pull/434)) by @mrexox

## 1.3.0 (2023-02-22)

- fix: Use correct branch for {push_files} template ([#429](https://github.com/evilmartians/lefthook/pull/429)) by @mrexox
- feature: Skip unstaged changes for pre-commit hook ([#402](https://github.com/evilmartians/lefthook/pull/402)) by @mrexox

## 1.2.9 (2023-02-13)

- fix: memory leak dependency ([#426](https://github.com/evilmartians/lefthook/pull/426)) by @strpc

## 1.2.8 (2023-01-23)

- fix: Don't join info path with root ([#418](https://github.com/evilmartians/lefthook/pull/418)) by @mrexox

## 1.2.7 (2023-01-10)

- fix: Make info dir when it is absent ([#414](https://github.com/evilmartians/lefthook/pull/414)) by @sato11
- deps: bump github.com/mattn/go-isatty from 0.0.16 to 0.0.17 ([#409](https://github.com/evilmartians/lefthook/pull/409)) by @dependabot
- deps: bump github.com/briandowns/spinner from 1.19.0 to 1.20.0 ([#406](https://github.com/evilmartians/lefthook/pull/406)) by @dependabot
- fix: Double quote eval statements with $dir ([#404](https://github.com/evilmartians/lefthook/pull/404)) by @jrfoell
- chore: Add PR template ([#401](https://github.com/evilmartians/lefthook/pull/401)) by @mrexox
- chore: Fix yml syntax missing colon ([#399](https://github.com/evilmartians/lefthook/pull/399)) by @aaronkelton

## 1.2.6 (2022-12-14)

- feature: Allow following output ([#397](https://github.com/evilmartians/lefthook/pull/397)) by @mrexox
- fix: Remove quotes for rc in template ([#398](https://github.com/evilmartians/lefthook/pull/398)) by @mrexox

## 1.2.5 (2022-12-13)

- feature: Add an option to disable spinner ([#396](https://github.com/evilmartians/lefthook/pull/396)) by @mrexox
- feature: Use pnpm before npx ([#393](https://github.com/evilmartians/lefthook/pull/393)) by @mrexox
- chore: Use lipgloss for output ([#395](https://github.com/evilmartians/lefthook/pull/395)) by @mrexox

## 1.2.4 (2022-12-05)

- feature: Allow providing rc file ([PR #392](https://github.com/evilmartians/lefthook/pull/392) by @mrexox)

## 1.2.3 (2022-11-30)

- feature: Expand env variables ([PR #391](https://github.com/evilmartians/lefthook/pull/391) by @mrexox)
- deps: Update important dependencies ([PR #389](https://github.com/evilmartians/lefthook/pull/389) by @mrexox)

## 1.2.2 (2022-11-23)

- chore: Add FreeBSD OS to packages ([PR #377](https://github.com/evilmartians/lefthook/pull/377) by @mrexox)
- feature: Skip based on branch name and allow global skip rules ([PR #376](https://github.com/evilmartians/lefthook/pull/376) by @mrexox)
- fix: Omit LFS output unless it is required ([PR #373](https://github.com/evilmartians/lefthook/pull/373) by @mrexox)

## 1.2.1 (2022-11-17)

- fix: Remove quoting for scripts ([PR #371](https://github.com/evilmartians/lefthook/pull/371) by @stonesbg + @mrexox)
- fix: remove lefthook.checksum on uninstall ([PR #370](https://github.com/evilmartians/lefthook/pull/370) by @JuliusHenke)
- fix: Print prepare-commit-msg hook if it exists in config ([PR #368](https://github.com/evilmartians/lefthook/pull/368) by @mrexox)
- fix: Allow changing refs for remote ([PR #363](https://github.com/evilmartians/lefthook/pull/363) by @mrexox)

## 1.2.0 (2022-11-7)

- fix: Full support for interactive commands and scripts ([PR #352](https://github.com/evilmartians/lefthook/pull/352) by @mrexox)
- chore: Remove deprecated config options ([PR #351](https://github.com/evilmartians/lefthook/pull/351) by @mrexox)
- feature: Add remote config support ([PR #343](https://github.com/evilmartians/lefthook/pull/343) by @oatovar and @mrexox)

## 1.1.4 (2022-11-1)

- feature: Add `LEFTHOOK_VERBOSE` env ([PR #346](https://github.com/evilmartians/lefthook/pull/346) by @mrexox)

## 1.1.3 (2022-10-15)

- ci: Fix snapcraft trying to create dirs in parallel by @mrexox
- feature: Allow setting env vars ([PR #337](https://github.com/evilmartians/lefthook/pull/337) by @mrexox)
- feature: Show current running command and script name(s) ([PR #338](https://github.com/evilmartians/lefthook/pull/338) by @mrexox)
- feature: Exclude by command names too ([PR #335](https://github.com/evilmartians/lefthook/pull/335) by @mrexox)
- fix: Don't uninstall lefthook.yml and lefthook-local.yml by default ([PR #334](https://github.com/evilmartians/lefthook/pull/334) by @mrexox)
- fix: Fixing typo in gemspec ([PR #333](https://github.com/evilmartians/lefthook/pull/333) by @kerrizor)

## 1.1.2 (2022-10-10)

- Fix regression from #314 (chmod missed fix) ([PR #330](https://github.com/evilmartians/lefthook/pull/330) by @ariccio)
- Pass stdin by default ([PR #324](https://github.com/evilmartians/lefthook/pull/324) by @mrexox)

## 1.1.1 (2022-08-22)

- Quote path to script ([PR #321](https://github.com/evilmartians/lefthook/pull/321) by @mrexox)

## 1.1.0 (2022-08-13)

- Add goreleaser action ([PR #307](https://github.com/evilmartians/lefthook/pull/307) by @mrexox)
- Windows escaping issues ([PR #314](https://github.com/evilmartians/lefthook/pull/314) by @mrexox)
- Check for lefthook.bat in hook template ([PR #316](https://github.com/evilmartians/lefthook/pull/316) by @mrexox)
- Update node.md docs ([PR #312](https://github.com/evilmartians/lefthook/pull/312) by @fantua)
- Move postinstall script to main lefthook NPM package ([PR #311](https://github.com/evilmartians/lefthook/pull/311) by @mrexox)
- Allow suppressing execution output ([PR #309](https://github.com/evilmartians/lefthook/pull/309) by @mrexox)
- Update dependencies ([PR #308](https://github.com/evilmartians/lefthook/pull/308) by @mrexox)
- Add support for Git LFS ([PR #306](https://github.com/evilmartians/lefthook/pull/306) by @mrexox)
- Bump Go version to 1.19 ([PR #305](https://github.com/evilmartians/lefthook/pull/305) by @mrexox)
- Add tests on runner ([PR #304](https://github.com/evilmartians/lefthook/pull/304) by @mrexox)
- Add fail text option ([PR #301](https://github.com/evilmartians/lefthook/pull/301) by @mrexox)
- Store lefthook checksum in non-hook file ([PR #280](https://github.com/evilmartians/lefthook/pull/280) by @mrexox)

## 1.0.5 (2022-07-19)

- Submodules issue ([PR #300](https://github.com/evilmartians/lefthook/pull/300) by @mrexox)
- Remove rspec tests ([PR #299](https://github.com/evilmartians/lefthook/pull/299) by @mrexox)
- Add `when "mingw" then "windows"` case ([PR #297](https://github.com/evilmartians/lefthook/pull/297) by @ariccio)
- Define security policy and contact method ([PR #293](https://github.com/evilmartians/lefthook/pull/293) by @Envek)

# 1.0.4 (2022-06-27)

- Support skipping on rebase ([PR #289](https://github.com/evilmartians/lefthook/pull/289) by @mrexox)

# 1.0.3 (2022-06-25)

- Fix NPM package
- Update email information

# 1.0.2 (2022-06-24)

- Bring auto install back ([PR #286](https://github.com/evilmartians/lefthook/pull/286) by @mrexox)
- Move main.go to root ([PR #285](https://github.com/evilmartians/lefthook/pull/285) by @mrexox)
- Panic on commands structure misuse ([PR #284](https://github.com/evilmartians/lefthook/pull/284) by @mrexox)
- Split npm package by os and cpu ([PR #281](https://github.com/evilmartians/lefthook/pull/281) by @mrexox)

# 1.0.1 (2022-06-20) Ruby gem and NPM package fix

- Fix folders structure for `@evilmartians/lefthook` and `@evilmartians/lefthook-installer` packages
- Fix folders structure for `lefthook` gem

# 1.0.0 (2022-06-19)

- Refactoring ([PR #275](https://github.com/evilmartians/lefthook/pull/275) by @mrexox, @skryukov, @markovichecha)
- Replace deprecated `File.exists?` with `exist?` for Ruby script ([PR #263](https://github.com/evilmartians/lefthook/pull/263) by @pocke)

# 0.8.0 (2022-06-07)

- Allow skipping hooks in certain git states: merge and/or rebase ([PR #173](https://github.com/evilmartians/lefthook/pull/173) by @DmitryTsepelev)
- NPM: installer package that downloads the required binaries during installation ([PR #188](https://github.com/evilmartians/lefthook/pull/188) by @aminya, [PR #273](https://github.com/evilmartians/lefthook/pull/273) by @Envek)
- Add ability to skip summary output. Also support a `LEFTHOOK_QUIET` env variable ([PR #187](https://github.com/evilmartians/lefthook/pull/187) by @washtubs)
- Make filename globs case-insensitive ([PR #196](https://github.com/evilmartians/lefthook/pull/196) by @skryukov)
- Fix lefthook binary extension on Windows ([PR #188](https://github.com/evilmartians/lefthook/pull/188) by @aminya)
- Stop building 32-bit binaries for releases due to low usage (@Envek)
- Allow lefthook to work when node_modules is not in root folder for npx ([PR #224](https://github.com/evilmartians/lefthook/pull/224) by @spearmootz)
- Fix unreachable conditional in hook template ([PR #242](https://github.com/evilmartians/lefthook/pull/242) by @dannobytes)
- Add cpu arch and os arch to lefthook's filepath in hook template ([PR #260](https://github.com/evilmartians/lefthook/pull/260) by @rmachado-studocu)

# 0.7.7 (2021-10-02)

- Fix incorrect npx command in git hook script template ([PR #236](https://github.com/evilmartians/lefthook/pull/236)) @PikachuEXE
- Update project URLs in NPM package.json ([PR #235](https://github.com/evilmartians/lefthook/pull/235)) @PikachuEXE
- Pass all arguments to downstream hooks ([PR #231](https://github.com/evilmartians/lefthook/pull/231)) @pablobirukov
- Allows lefthook to work when node_modules is not in root folder for npx ([PR #224](https://github.com/evilmartians/lefthook/pull/224)) @spearmootz
- Do not initialize git config on `help` and `version` commands ([PR #209](https://github.com/evilmartians/lefthook/pull/209)) @pwinckles
- node: fix postinstall: process.cwd is a function and should be called @Envek

# 0.7.6 (2021-06-02)

- Fix lefthook binary extension on Windows. @aminya
- [PR #193](https://github.com/evilmartians/lefthook/pull/193) Fix path for searching npm-installed binary when in worktree. @Envek
- NPM, RPM, and DEB packaging fixes. @Envek

# 0.7.5 (2021-05-14)

- [PR #179](https://github.com/evilmartians/lefthook/pull/179) Fix running on Windows under MSYS and MINGW64 when run from Ruby gem or JS npm package. @akiver, @Envek
- [PR #177](https://github.com/evilmartians/lefthook/pull/177) Support non-default git hooks path. @charlie-wasp
- [PR #182](https://github.com/evilmartians/lefthook/pull/182) Support git workspaces and submodules. @skryukov
- [PR #184](https://github.com/evilmartians/lefthook/pull/184) Rewrite npm's scripts in JavaScript to support running on Windows without `sh`. @aminya

# 0.7.4 (2021-04-30)

- [PR](https://github.com/evilmartians/lefthook/pull/171) Improve check for installed git @DmitryTsepelev
- [PR](https://github.com/evilmartians/lefthook/pull/169) Create .git/hooks directory when it does not exist @DmitryTsepelev

# 0.7.3 (2021-04-23)

- [PR](https://github.com/evilmartians/lefthook/pull/168) Package versions for all architectures (x86_64, ARM64, x86) into Ruby gem and NPM package @Envek
- [PR](https://github.com/evilmartians/lefthook/pull/167) Fix golang 15+ build @skryukov

# 0.7.2 (2020-02-02)

- [PR](https://github.com/evilmartians/lefthook/pull/126) Feature multiple extends. Thanks @Evilweed

- [PR](https://github.com/evilmartians/lefthook/pull/124) Fix `npx` when only `yarn` exists. Thanks @dotterian

- [PR](https://github.com/evilmartians/lefthook/pull/116) Fix use '-h' for robust lefthook. Thanks @fahrinh

# 0.7.1 (2020-02-02)

- [PR](https://github.com/evilmartians/lefthook/pull/108) Fix `sh` dependency on windows when executing `git`. Thanks @lionskape

- [PR](https://github.com/evilmartians/lefthook/pull/103) Add possibility for using `yaml` and `yml` extension for config. Thanks @rbUUbr

# 0.7.0 (2019-12-14)

- [PR](https://github.com/evilmartians/lefthook/pull/98) Support relative roots for monorepos. Thanks @jsmestad

# 0.6.7 (2019-12-14)

- [Commit](https://github.com/evilmartians/lefthook/commit/e898b5c8ba56c4d6f29a4d1f433baa1779a0845b)
Skip before executing command

- [PR](https://github.com/evilmartians/lefthook/pull/94) Add option --keep-config. Thanks @justinasposiunas

- [Commit](https://github.com/evilmartians/lefthook/commit/d79a3a46e7d1ee709b97e97f823bfd27e9466eff)
Check if shell is non interactive

# 0.6.6 (2019-12-03)

- [PR](https://github.com/evilmartians/lefthook/pull/94) Use eval instead of exec; Enable tty for the shell. Thanks @ssnickolay

# 0.6.5 (2019-11-15)

- [PR](https://github.com/evilmartians/lefthook/pull/89) Add support for git-worktree. Thanks @f440

- [Commit](https://github.com/evilmartians/lefthook/commit/48702a0806d2b2eab13636ba56b0e0b99f346f1c)
Commands and Scripts now can catch Stdin

- [Commit](https://github.com/evilmartians/lefthook/commit/9a226842292ff1dda0f2273b66a0799988aa5289)
Add partial support for monorepos and command execution not from project root

# 0.6.4 (2019-11-08)

- [PR](https://github.com/evilmartians/lefthook/pull/84) Fix return value from shell exit. Thanks @HaiD84

- [PR](https://github.com/evilmartians/lefthook/pull/86) Support postinstall script for npm installation for monorepos. Thanks @sHooKDT

- [PR](https://github.com/evilmartians/lefthook/pull/82) Now relative path to scripts supported. Thanks @AlexeyMatskevich

- [Commit](https://github.com/evilmartians/lefthook/pull/80/commits/1a4b0ee155eb66ae6f3c365164012bee9332605a)
Option `extends` for top level config added. Now you can merge some settings from different places:
```yml
extends: $HOME/work/lefthook-extend.yml
```

- [Commit](https://github.com/evilmartians/lefthook/commit/83cf818106dbf222ea33ba86aafce8f30d7cb5a9)
Add examples to generated lefthook.yml

## 0.6.3 (2019-07-15)

- [Commit](https://github.com/evilmartians/lefthook/commit/0426936f48f248221126f15619932b0dc8c54d7a) Add `-a` means `aggressive` strategy for `install` command
```bash
lefthook install -a # clear .git/hooks dir and reinstall lefthook hooks
```

- [Commit](https://github.com/evilmartians/lefthook/commit/5efb0677a4a9ec1728d3cf1a083075e23315a796) Add Lefthook version indicator for commands and script execution

- [Commit](https://github.com/evilmartians/lefthook/commit/8b55d91eed46643a1674bd4ad96fa211a177e159) Remove `npx` as dependency from node wrapper

Now we will call directly binary from `./node_modules`

- [Commit](https://github.com/evilmartians/lefthook/commit/76ffed4c698bc074984e91f5610c0b98784bd10b) Add `-f` means `force` strategy for `install` command

```bash
lefthook install -f # reinstall lefthook hooks without sync info check
```

- PR [#27](https://github.com/evilmartians/lefthook/pull/27) Move LEFTHOOK env check in hooks files

Now if LEFTHOOK=0 we will not call the binary file

- PR [#26](https://github.com/evilmartians/lefthook/pull/26) + [commit](https://github.com/evilmartians/lefthook/commit/afd67f94631a10975209ed4c5fabc763f44280eb) Add `{push_files}` shortcut

Add shortcut `{push_files}`

```
pre-commit:
  commands:
    rubocop:
      run: rubocop {push_files}
```
It same as:
```
pre-commit:
  commands:
    rubocop:
      files: git diff --name-only HEAD @{push} || git diff --name-only HEAD master
      run: rubocop {push_files}
```

- [Commit](https://github.com/evilmartians/lefthook/commit/af087b032a14952aa1dd235a3d0b5a51bc760a10) Add `min_version` option

You can mark your config for minimum Lefthook version:
```
min_version: 0.6.1
```

## 0.6.0 (2019-07-10)

- PR [#24](https://github.com/palkan/logidze/pull/110) Wrap `run` command in shell context.

Now in `run` option available `sh` syntax.

```
pre-commit:
  commands:
    bashed:
      run: rubocop -a && git add
```
Will be executed in this way:
```
sh -c "rubocop -a && git add"
```

- PR [#23](https://github.com/evilmartians/lefthook/pull/24) Search Lefthook in Gemfile.

Now it's possible to use Lefthook from Gemfile.

```ruby
# Gemfile

gem 'lefthook'
```
