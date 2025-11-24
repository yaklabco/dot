<a name="unreleased"></a>
## [0.6.0](https://github.com/yaklabco/dot/compare/dot-v0.5.0...dot-v0.6.0) (2025-11-24)


### ⚠ BREAKING CHANGES

* **cli:** Remove automatic glob mode detection to eliminate ambiguous behavior. Users must now provide explicit package names when adopting multiple files.
* **adopt:** Adopt now creates flat package structure
* **packages:** Package structure requirements changed.

### Features

* **adapters,planner,cli:** add GitHub CLI auth, backup system, and conflict resolution ([ab4a430](https://github.com/yaklabco/dot/commit/ab4a430ed802e3381555a4d291a9f0e53d5f09f5))
* **adapters:** integrate GitHub CLI authentication ([7f81270](https://github.com/yaklabco/dot/commit/7f8127092f290a76be8499f0b416f9218fd633f4))
* **adopt:** add auto-naming and glob expansion modes ([5951367](https://github.com/yaklabco/dot/commit/595136751fefd939240d14d6b7928c3b98b0c88c))
* **adopt:** add file categorization system ([19942b6](https://github.com/yaklabco/dot/commit/19942b651027b64e6badd6bf698341a2cd86667e))
* **adopt:** add file discovery and helpers ([56fd638](https://github.com/yaklabco/dot/commit/56fd638e91311f8f0358e1e41071e4a3f25a5fd1))
* **adopt:** add interactive file selector ([53cded5](https://github.com/yaklabco/dot/commit/53cded57328175082e400c1e3e267bb480f083ed))
* **adopt:** add interactive session management ([836b49b](https://github.com/yaklabco/dot/commit/836b49b55884919e1ab8f46100cbd9cfe85112b0))
* **api,cli:** implement Client API and CLI infrastructure ([5e668ce](https://github.com/yaklabco/dot/commit/5e668ce833db7d44b1ade6bee6a8846e875bb096))
* **api:** add comprehensive tests and documentation ([8f23f6a](https://github.com/yaklabco/dot/commit/8f23f6ad8d7b690b853d49b3009dfd7c22499a1c))
* **api:** add depth calculation and directory skip logic ([715f0c6](https://github.com/yaklabco/dot/commit/715f0c69df631333c9f820e15023e52e85f2af1b))
* **api:** add DoctorWithScan for explicit scan configuration ([4bae524](https://github.com/yaklabco/dot/commit/4bae524e03cef7487d8e52b742b5d262a251f784))
* **api:** add foundational types for Client API ([2361b1c](https://github.com/yaklabco/dot/commit/2361b1c3e7dedeafb7cd163c877cbb241be433f9))
* **api:** define Client interface for public API ([eb83b8a](https://github.com/yaklabco/dot/commit/eb83b8a653442c99c731839fd725b1d8823207b0))
* **api:** implement all stubbed features with intelligent incremental operations ([b2b9dd3](https://github.com/yaklabco/dot/commit/b2b9dd32adbda29fb1a175a084c841e8234f6ef7))
* **api:** implement Client with Manage operation ([78bc3c8](https://github.com/yaklabco/dot/commit/78bc3c8da800ea0ba08c12856f8ffc3630ac45b3))
* **api:** implement directory extraction and link set optimization ([5582e08](https://github.com/yaklabco/dot/commit/5582e081e1d255ea698d3898ce8f4b64e8611a5b))
* **api:** implement incremental remanage with hash-based change detection ([9ee9fae](https://github.com/yaklabco/dot/commit/9ee9fae775e26e7334f69fc47761096095d58874))
* **api:** implement link count extraction from plan ([c30a0db](https://github.com/yaklabco/dot/commit/c30a0dbdd16bb51d8f11b2a84b5714fe02b0a020))
* **api:** implement orphaned link detection and link count tracking ([f892717](https://github.com/yaklabco/dot/commit/f892717b04e42cf4cd3137aa6106f684421facb8))
* **api:** implement Public Library API ([72d607b](https://github.com/yaklabco/dot/commit/72d607b91abea7b55714849c0d8057c17965a192))
* **api:** implement Unmanage, Remanage, and Adopt operations ([2ba4e93](https://github.com/yaklabco/dot/commit/2ba4e935a3fe821c6b37a6a80a1895d481abdb0f))
* **api:** integrate backup system with manage service ([a1c71c7](https://github.com/yaklabco/dot/commit/a1c71c7ddd06a20a7e85595e0aa3afd64ba99c23))
* **api:** update Doctor API to accept ScanConfig parameter ([8a2f166](https://github.com/yaklabco/dot/commit/8a2f166a66de7f7978d163a663d1c65fb1f5f7b7))
* **api:** update Doctor API to accept ScanConfig parameter ([dbfe2c3](https://github.com/yaklabco/dot/commit/dbfe2c35c2b299f7fe25e431552cf87dbbd8069d))
* **api:** wire up orphaned link detection with safety limits ([2fc21e2](https://github.com/yaklabco/dot/commit/2fc21e24d9c637a67d4d53b38b2945c928ceaf5f))
* **bootstrap:** add configuration schema and loader ([859db98](https://github.com/yaklabco/dot/commit/859db98833fc78e647d8dc665497e6e3a4c20951))
* **bootstrap:** implement bootstrap config generator ([4765c01](https://github.com/yaklabco/dot/commit/4765c0124beba88f5fb2a65379b14b4e2a2d28a5))
* **cli:** add batch mode and async version check ([a954f95](https://github.com/yaklabco/dot/commit/a954f9553b69ab5b362e1b04c5f255ddcaa35874))
* **cli:** add color helper for dynamic color detection ([b2d19b9](https://github.com/yaklabco/dot/commit/b2d19b9ffee4646dedd3e546ee6c5eb8302e055d))
* **cli:** add command aliases for config ([7ed870b](https://github.com/yaklabco/dot/commit/7ed870b3c3c7bb1f1cb004aa1a4ade9e307b7ce0))
* **cli:** add comprehensive tab-completion and unmanage restoration ([d12b458](https://github.com/yaklabco/dot/commit/d12b458fd71d662d487d79cfffb21705fef92d54))
* **cli:** add config command for XDG configuration management ([bf3c981](https://github.com/yaklabco/dot/commit/bf3c981583c12bd8343640d927339bb17ad9d901))
* **cli:** add interactive package selector and terminal detection ([0ec25a4](https://github.com/yaklabco/dot/commit/0ec25a45609ef03de6ea72d25f27067297d037f6))
* **cli:** add interactive prompt utilities ([bee3ec2](https://github.com/yaklabco/dot/commit/bee3ec24b8d6ce144f2ee508e97a1d759792d6fd))
* **cli:** add list and text formatting utilities ([7269be6](https://github.com/yaklabco/dot/commit/7269be62008fd78b81d2e4efbf7b7256231745db))
* **cli:** add muted colorization to doctor and unmanage commands ([07b45cf](https://github.com/yaklabco/dot/commit/07b45cf16c28e0187690d8201d6d991aa0b0648f))
* **cli:** add pluralization helpers for output formatting ([fe110bc](https://github.com/yaklabco/dot/commit/fe110bc72081d496f47ab03a92c7f76fdcf59c55))
* **cli:** add profiling and diagnostics support ([f5cc629](https://github.com/yaklabco/dot/commit/f5cc629e8325c86da2bd0a52ff036a72828a97d5))
* **cli:** add progress tracking using go-pretty ([44e23d0](https://github.com/yaklabco/dot/commit/44e23d0a17807a260df78d1bba13350f79ad4e33))
* **cli:** add scan control flags to doctor command ([915bf3a](https://github.com/yaklabco/dot/commit/915bf3abb84495274cbc0de33038252b5fa46cdc))
* **cli:** add table infrastructure using go-pretty ([ee9d71a](https://github.com/yaklabco/dot/commit/ee9d71aa225acfe85360921e0edf9957b42e02d1))
* **cli:** comprehensive UX improvements and bug fixes ([9f24a63](https://github.com/yaklabco/dot/commit/9f24a6314ce645145a230817445864a3580da096))
* **cli:** display usage help on invalid flags and arguments ([eaac30c](https://github.com/yaklabco/dot/commit/eaac30c17dc70facb55d360d306e7de477f27169))
* **client:** integrate CloneService into Client facade ([9cd27d6](https://github.com/yaklabco/dot/commit/9cd27d6c16ca34348df2c23ad993cc70bfb3d38a))
* **cli:** implement CLI infrastructure with core commands ([9ebf4ff](https://github.com/yaklabco/dot/commit/9ebf4ff54f4a963f1519a2ccfc3e161426aa2086))
* **cli:** implement command handlers for manage, unmanage, remanage, adopt ([7b14ce9](https://github.com/yaklabco/dot/commit/7b14ce947427891b55bf562a139ead645c9b8720))
* **cli:** implement doctor command for health checks ([546f162](https://github.com/yaklabco/dot/commit/546f1625d609c0c1a8e94d6f34ebe17121aa4d78))
* **cli:** implement dot clone command ([dadeabd](https://github.com/yaklabco/dot/commit/dadeabda3ab32a3ff042bdd3d0c6faca232e094a))
* **cli:** implement dot upgrade command with comprehensive testing ([e406af3](https://github.com/yaklabco/dot/commit/e406af3a29365fad1115cb70432b3441d236fefa))
* **cli:** implement error formatting foundation for ([a9308b5](https://github.com/yaklabco/dot/commit/a9308b518ee5e16485f6d303098f9030fab9cfe3))
* **cli:** implement error handling and user experience ([938fe5a](https://github.com/yaklabco/dot/commit/938fe5a698ce91ff7bc88b32609e43c0f6ce8af6))
* **cli:** implement help system with examples and completion ([90ed120](https://github.com/yaklabco/dot/commit/90ed1205bf63076541f78e2833773545518e7032))
* **cli:** implement list command for package inventory ([cd4dac8](https://github.com/yaklabco/dot/commit/cd4dac8cf9dc1276d2573d2c285a998e5f4ed81e))
* **cli:** implement output renderer infrastructure ([39d07c2](https://github.com/yaklabco/dot/commit/39d07c2d29932199ed725103747dcb7b54037c27))
* **cli:** implement progress indicators for operation feedback ([fcb0ce5](https://github.com/yaklabco/dot/commit/fcb0ce5a479311b4a5de6a469f5e4ed39ec08411))
* **cli:** implement query commands with flexible output rendering ([271e5dd](https://github.com/yaklabco/dot/commit/271e5ddc6a2bf003bb9472fea0cc3ac83b24c2ba))
* **cli:** implement status command for installation state inspection ([ad3e81a](https://github.com/yaklabco/dot/commit/ad3e81a98b0bc0c636ad36c17bb1280719d6fcad))
* **cli:** implement terminal styling and layout system ([17519cb](https://github.com/yaklabco/dot/commit/17519cb79d55955889df46d687b334a808742aa5))
* **cli:** implement upgrade command with automatic version checking ([2724efd](https://github.com/yaklabco/dot/commit/2724efd8f3f85ce3960b0829794da7041447e99b))
* **cli:** implement UX polish with output formatting ([030df75](https://github.com/yaklabco/dot/commit/030df7556f2a665479db4f56680c949095215028))
* **cli:** improve package selection UI with colors and layout ([8aa7f05](https://github.com/yaklabco/dot/commit/8aa7f053d5fe9e38721b778d3b79b82fb1d88e13))
* **cli:** integrate startup version check into root command ([d75814d](https://github.com/yaklabco/dot/commit/d75814d1aab2bd32c3bed75112f6bfafd5d16050))
* **cli:** modernize UI with professional rendering and enhanced UX ([786fe17](https://github.com/yaklabco/dot/commit/786fe17a17a5c739257b7f62d483a43a2d9457fd))
* **cli:** refactor table rendering with go-pretty for professional UX ([2e98aaa](https://github.com/yaklabco/dot/commit/2e98aaa189d36bb9ea7f10d6dc4417d81ccc7ea7))
* **cli:** show complete operation breakdown in table summary ([8a1c34f](https://github.com/yaklabco/dot/commit/8a1c34fe5ba2fe5fbd9cc5c49ab890948a372bd2))
* **cli:** simplify adopt command behavior ([80be455](https://github.com/yaklabco/dot/commit/80be4557b2dc8b4e0c3bd1514517869c30de1a0d))
* **clone:** add bootstrap generation subcommand ([75d2a93](https://github.com/yaklabco/dot/commit/75d2a934f0ec5c6bb44474a0cfc611adb6fec668))
* **clone:** add hierarchical package directory resolution and self-management prevention ([946c022](https://github.com/yaklabco/dot/commit/946c022bf38f415376aba49d3c2eb652bf698914))
* **clone:** derive directory from repository name like git clone ([cc12200](https://github.com/yaklabco/dot/commit/cc12200db9f52874e5d0e7660f14ba465c4e34fa))
* **clone:** implement clone bootstrap subcommand ([d855d6c](https://github.com/yaklabco/dot/commit/d855d6c61c9f9072d0b34e29a5f39e2ba8375178))
* **clone:** implement CloneService orchestrator ([f2d00fe](https://github.com/yaklabco/dot/commit/f2d00fef4d3011c92054e3ff8e0c4086e2eb9ba0))
* **clone:** implement dot clone command with bootstrap configuration ([d896c2d](https://github.com/yaklabco/dot/commit/d896c2dcf05c9a875618904b8f085858f3e5f774))
* **clone:** improve branch detection and profile filtering ([5d1d2e4](https://github.com/yaklabco/dot/commit/5d1d2e495c8d9b456a2631a4ae1a27fbb064141c))
* **commands:** add secrets detection helpers ([0cd3673](https://github.com/yaklabco/dot/commit/0cd36736d2bbf81076adae1fc068672028fd8483))
* **config:** add backup and overwrite configuration options ([4809f74](https://github.com/yaklabco/dot/commit/4809f74b61ce2cce42c6e7be5a7a2ecfb61a2227))
* **config:** add Config struct with validation ([7d2ac49](https://github.com/yaklabco/dot/commit/7d2ac49a6cf3a2a841e8f0b92d58e58cdb8c785b))
* **config:** add configuration operations and upgrade ([729d086](https://github.com/yaklabco/dot/commit/729d0866066d41bc487967d472b4ebb69a9dd1e6))
* **config:** add network configuration support ([c82ec6c](https://github.com/yaklabco/dot/commit/c82ec6cebeec0ff736cb3c90ad4564226eb4de3b))
* **config:** add update configuration for package manager and version checking ([b535f75](https://github.com/yaklabco/dot/commit/b535f755c26f1c5b30d38386d18fee6c36c28daa))
* **config:** add validation for network timeout fields ([985f9ef](https://github.com/yaklabco/dot/commit/985f9ef30407340e72ce2fcb82ba86e6aa378d03))
* **config:** implement extended configuration infrastructure ([cae9c95](https://github.com/yaklabco/dot/commit/cae9c95642f26ccffd5d2dee93bca5e43cbee687))
* **config:** implement JSON marshal strategy ([2b8c98a](https://github.com/yaklabco/dot/commit/2b8c98a6b881a94632b5e671ccca9e5ec7a30c5a))
* **config:** implement TOML marshal strategy ([ef719e0](https://github.com/yaklabco/dot/commit/ef719e0b837c91caaf06237fff05186adebff45b))
* **config:** implement YAML marshal strategy ([d3a88d9](https://github.com/yaklabco/dot/commit/d3a88d90e9326c4da9eba8ea3f915368ca8575c8))
* **config:** wire backup directory through system ([1d3ca5a](https://github.com/yaklabco/dot/commit/1d3ca5a2b3ab495a4691b55d8548befa13f62c18))
* **config:** wire up table_style configuration to all commands ([cf19892](https://github.com/yaklabco/dot/commit/cf19892c2a996271129a141c532c5c7694dc430f))
* **doctor:** add interactive fix and ignore capabilities ([b960b48](https://github.com/yaklabco/dot/commit/b960b48173c943e8c86f8ffc732d5c3b9c1879f5))
* **doctor:** add interactive triage and health checking system ([a4f7ba5](https://github.com/yaklabco/dot/commit/a4f7ba55c14d73ec9637c566375e8d67594dba6a))
* **doctor:** add interactive triage for orphaned symlinks ([7b8cf28](https://github.com/yaklabco/dot/commit/7b8cf28c831cff21c4d2a09a635e49c645401c84))
* **doctor:** add secrets detection module ([4d30c6f](https://github.com/yaklabco/dot/commit/4d30c6f3e2d47ca2e02b9217c7c38b6d0547071c))
* **doctor:** add triage flag for interactive orphan management ([e345625](https://github.com/yaklabco/dot/commit/e345625236bef322679d739f14f3010963950f5c))
* **doctor:** enable scoped orphan scanning by default and detect broken unmanaged links ([3456b92](https://github.com/yaklabco/dot/commit/3456b92806ac16be8ce4d0940524a630933e90ed))
* **doctor:** unify health checking across list and doctor ([18b955c](https://github.com/yaklabco/dot/commit/18b955ca77006610b4cc1b12af08d766c2103de6))
* **domain:** add chainable Result methods ([7eead68](https://github.com/yaklabco/dot/commit/7eead6878b3c1c8e1c65c3c984d8afd1e8fd30c5))
* **domain:** add FileDelete operation and fix FileBackup permissions ([9c02037](https://github.com/yaklabco/dot/commit/9c0203731e597c9ce64b8d0aeaf5f5d88ada75d6))
* **domain:** add Lstat method to FS interface ([f84b393](https://github.com/yaklabco/dot/commit/f84b3936de4f42b89990feb3b164a12154c6605d))
* **domain:** add package-operation mapping to Plan ([5dbbb29](https://github.com/yaklabco/dot/commit/5dbbb294e914fd354ea419a56c14e0523397b18a))
* **dot:** add bootstrap generation to Client facade ([3c97a8e](https://github.com/yaklabco/dot/commit/3c97a8e1778df15828e09e7ad6e9c27a54bbc801))
* **dot:** add ScanConfig types for orphaned link detection ([361baa4](https://github.com/yaklabco/dot/commit/361baa4c850fb45607df10d820aeaa4c437c30b5))
* **errors:** add clone-specific error types ([125b02a](https://github.com/yaklabco/dot/commit/125b02a0cec2ef3026ca31419eb84f47328c0eb2))
* **executor:** add metrics instrumentation wrapper ([761ce3b](https://github.com/yaklabco/dot/commit/761ce3b0f067ca71ffc4be742d7f68a1990e1e6c))
* **executor:** implement executor with two-phase commit ([241c50c](https://github.com/yaklabco/dot/commit/241c50c5c8d6202377e06c3bd3e4171bd5c48da9))
* **executor:** implement imperative shell with transaction safety ([d012c2e](https://github.com/yaklabco/dot/commit/d012c2e792ca619a44bfc501ff8a08c3a9176739))
* **executor:** implement parallel batch execution ([fb0d2e3](https://github.com/yaklabco/dot/commit/fb0d2e344270f55dc9608d535425cb05d1a1b4c4))
* **git:** add git cloning with authentication support ([b280b90](https://github.com/yaklabco/dot/commit/b280b90d9dd67843bd5cb29eb4ba4a1f64d9f9d1))
* **ignore:** add .dotignore file support ([ba502e2](https://github.com/yaklabco/dot/commit/ba502e2f60ffcbdd194334c569abc0be4f0de47a))
* **ignore:** extend default patterns for security ([5806da0](https://github.com/yaklabco/dot/commit/5806da09f5d716646350386db061c43671d63857))
* **list:** add --show-target flag to display target directory ([d26fb57](https://github.com/yaklabco/dot/commit/d26fb577216d7231ed0d2d80d5e394b618c0bb6d))
* **list:** add health status indicators to list output ([c6ed0ef](https://github.com/yaklabco/dot/commit/c6ed0ef86367a6359a1e542b3c84e3c122bdee30))
* **list:** align columns in text output format ([57b168e](https://github.com/yaklabco/dot/commit/57b168eedecb324aabf506f1a5582bb15a2dfe11))
* **manage,adopt:** warn about sensitive files ([6100543](https://github.com/yaklabco/dot/commit/6100543038c3bc23b61e430387d7d3cef8aeaf90))
* **manifest:** add backup tracking to package metadata ([3ab5c9f](https://github.com/yaklabco/dot/commit/3ab5c9fbdbe4a3e55d150d87d9654b5dd361fae6))
* **manifest:** add core manifest domain types ([eb3cb5a](https://github.com/yaklabco/dot/commit/eb3cb5aa583c573765a59468447b1e8dbd1c6942))
* **manifest:** add doctor state tracking and pattern categorization ([834f9d1](https://github.com/yaklabco/dot/commit/834f9d18d2043abc28c17aeb13bc34b1b6ccec13))
* **manifest:** add repository tracking support ([f962a75](https://github.com/yaklabco/dot/commit/f962a75a3c187621a24afa8f1fb3549a4abb66ed))
* **manifest:** define ManifestStore interface ([ef262b9](https://github.com/yaklabco/dot/commit/ef262b9558c0fae5e9c5fcff9007b7af71de4b79))
* **manifest:** implement content hashing for packages ([2def4bc](https://github.com/yaklabco/dot/commit/2def4bcbd769542622545d7917c3811d4e95f03f))
* **manifest:** implement FSManifestStore persistence ([097fc56](https://github.com/yaklabco/dot/commit/097fc562b89b53cd5a491770ef5317490fc23059))
* **manifest:** implement manifest validation ([09ff2c9](https://github.com/yaklabco/dot/commit/09ff2c9f13866b34e14182927e5197f3784a6b58))
* **operation:** add Execute and Rollback methods to operations ([eb319a4](https://github.com/yaklabco/dot/commit/eb319a435d387f9119b5039b9bdf257febe4432a))
* **packages:** enable package name mapping and improve CLI error handling ([b471ee0](https://github.com/yaklabco/dot/commit/b471ee0de9ee262d414927afa86c69edcb05977a))
* **packages:** enable package name to target directory mapping ([1bc4bc5](https://github.com/yaklabco/dot/commit/1bc4bc579e2fb2bdfad7f5b736dd7824777367f1))
* **pager:** add keyboard controls for interactive pagination ([07e1a42](https://github.com/yaklabco/dot/commit/07e1a429eb840e5e0fcbddeaab0696ba8382bb08))
* **pipeline:** add current state scanner for conflict detection ([4cae8ec](https://github.com/yaklabco/dot/commit/4cae8ec4327b3fc5d53307e76669eff9c5897b29))
* **pipeline:** enhance context cancellation handling in pipeline stages ([b438c5b](https://github.com/yaklabco/dot/commit/b438c5b562d615599b3787e93650129afb6a800e))
* **pipeline:** implement symlink pipeline with scanning, planning, resolution, and sorting stages ([43dd10d](https://github.com/yaklabco/dot/commit/43dd10d4319131ddbcd52dd0574b429e2020b560))
* **pipeline:** implement symlink pipeline with scanning, planning, resolution, and sorting stages ([dbb1bdb](https://github.com/yaklabco/dot/commit/dbb1bdb50ea161916962ecbd6f14e2c7cb832911))
* **pipeline:** surface conflicts and warnings in plan metadata ([38a3de1](https://github.com/yaklabco/dot/commit/38a3de1dcac2010a40b9a33bf5f4a291a8caefb2))
* **pipeline:** track package ownership in operation plans ([b940fdd](https://github.com/yaklabco/dot/commit/b940fdd902c59ceb76402cb58c7fbcfbc7a273d0))
* **pkg:** extract AdoptService from Client ([dd927d6](https://github.com/yaklabco/dot/commit/dd927d6bedd927c7f758eeb499c77d00b79c09b5))
* **pkg:** extract DoctorService from Client ([6837086](https://github.com/yaklabco/dot/commit/6837086b9e8db1d63b50c9a4fe00f5710cfaf71e))
* **pkg:** extract ManageService from Client ([ff46f77](https://github.com/yaklabco/dot/commit/ff46f7799a62f232a788208b75d149ddc731a0ba))
* **pkg:** extract ManifestService from Client ([24865dd](https://github.com/yaklabco/dot/commit/24865dd1778c81b0e5a7b80c0d8cb27a45ba0bb6))
* **pkg:** extract StatusService from Client ([9f0154f](https://github.com/yaklabco/dot/commit/9f0154fbe9ee4203ba03d0c01e3fc9672a7a9704))
* **pkg:** extract UnmanageService from Client ([756da92](https://github.com/yaklabco/dot/commit/756da9218bec8af95d8dd5365ec02d153e17a5eb))
* **planner:** implement backup and overwrite conflict policies ([542bb18](https://github.com/yaklabco/dot/commit/542bb18e45396a129c7fcd09b9f3a3c054f63488))
* **planner:** implement parallelization analysis ([52df44c](https://github.com/yaklabco/dot/commit/52df44c88fed63030dca46aa368f96c3d17d3362))
* **planner:** implement topological sort with cycle detection ([287ad8d](https://github.com/yaklabco/dot/commit/287ad8dddd3adc8a3be39e1974a830571541dd10))
* **release:** automate release workflow with integrated changelog generation ([118f8f1](https://github.com/yaklabco/dot/commit/118f8f1587e84d070eec4f29647cb4e1078650f0))
* **renderer:** add health column to table output ([75da9a7](https://github.com/yaklabco/dot/commit/75da9a76b31e06f1e9763596167d6dc0041c1ac6))
* **retry:** add exponential backoff retry utility ([c441ca9](https://github.com/yaklabco/dot/commit/c441ca942c86761140edc65de2bbfd0f0bbf0d39))
* **scanner:** add prompter and enhance tree scanning ([53dd7c1](https://github.com/yaklabco/dot/commit/53dd7c123221fb25361edfc71de3982b7acb723e))
* **status:** add health checking to status service ([a190c5a](https://github.com/yaklabco/dot/commit/a190c5ac9f6f7558859eb4c4527f10ecf5e17990))
* **testing:** enhance infrastructure with profiling and security ([524db43](https://github.com/yaklabco/dot/commit/524db4318570b049e78de01e1b3a58c164223a5a))
* **types:** add Status and PackageInfo types ([244eba8](https://github.com/yaklabco/dot/commit/244eba8039136d7fad2589eeab401911143d73d1))
* **ui:** add automatic pagination to dot doctor output ([821aa21](https://github.com/yaklabco/dot/commit/821aa211f63d8f5168c52e653724ac43a51d4e52))
* **ui:** add configuration toggle between modern and legacy table styles ([1be8615](https://github.com/yaklabco/dot/commit/1be86153010caaba7f62b0869f9d2fcec1f8f972))
* **unmanage:** add --all flag to unmanage all packages at once ([82cd17d](https://github.com/yaklabco/dot/commit/82cd17dd97ddceb79b231c13c1a0cae884148634))
* **updater:** add colorized output to update notification ([9d5b645](https://github.com/yaklabco/dot/commit/9d5b64574a41550ee6707e0c36feba1ec2bfb883))
* **updater:** add security validation for package managers ([05a7cc7](https://github.com/yaklabco/dot/commit/05a7cc74dadfea281983b635c8f315e816534dc1))
* **updater:** add version checker and package manager services ([66ba0c4](https://github.com/yaklabco/dot/commit/66ba0c488bfe5fa2bf2e0600cfc686cd4e894b6c))
* **updater:** implement startup version checking system ([c1ac265](https://github.com/yaklabco/dot/commit/c1ac2654484f2cba9097aadd548912667503d2b3))
* **updater:** improve version checking with better error handling ([8cae985](https://github.com/yaklabco/dot/commit/8cae985e0b5c8ccaee905b12fddd11a98d9efde6))


### Bug Fixes

* **adapters:** use correct HTTP Basic Auth format for tokens ([7cff7db](https://github.com/yaklabco/dot/commit/7cff7db531124401ba76e0ed3ff682a12038788c))
* **adopt:** prevent re-adoption of managed symlinks ([298efb6](https://github.com/yaklabco/dot/commit/298efb66356c1fcf3bdfbed7cd1157b4075c6126))
* **api:** address CodeRabbit feedback on ([710d06b](https://github.com/yaklabco/dot/commit/710d06bd1e8a32646b0d4dbfa7a1648468c91bf0))
* **api:** enforce depth and context limits in recursive orphan scanning ([22a33d7](https://github.com/yaklabco/dot/commit/22a33d7316107a96646168bd5f245a4a3034095e))
* **api:** improve error handling and test robustness ([f1c8d4c](https://github.com/yaklabco/dot/commit/f1c8d4c105bae3248b9632c1d9b2b1c1d5fba5e3))
* **api:** normalize paths for cross-platform link lookup ([6b83e40](https://github.com/yaklabco/dot/commit/6b83e40894a3dc6c28968fd9b8fb33620ac56aa0))
* **api:** use configured skip patterns in recursive orphan scanning ([a2a539d](https://github.com/yaklabco/dot/commit/a2a539d9fcdbb2a2b7e351e9fe4fc5d0c9d86bd0))
* **api:** use package-operation mapping for accurate manifest tracking ([39a833d](https://github.com/yaklabco/dot/commit/39a833d822ba3950b53f0b73e7e37177770fe319))
* **bootstrap:** correct documentation URL in generated config header ([85e4033](https://github.com/yaklabco/dot/commit/85e4033ff8a9707d869df06e6fc018c2666db7ac))
* **bootstrap:** implement manifest parsing for from-manifest flag ([5132fac](https://github.com/yaklabco/dot/commit/5132facacb3a03bff97a2e6a6b5d7727cb5e5c32))
* **bootstrap:** use installed parameter in YAML comments ([1699ac5](https://github.com/yaklabco/dot/commit/1699ac52f9531403003b03f7aec009330d2e5733))
* **ci:** remove path to golangci-lint ([330e891](https://github.com/yaklabco/dot/commit/330e891a0e08ac16cd1b77bfbe77a387d145a2ac))
* **ci:** remove path to golangci-lint ([0fea1e9](https://github.com/yaklabco/dot/commit/0fea1e96b0d01e0642473adbe172a1cc935da9ea))
* **cli:** add error templates for checkpoint and not implemented errors ([c689f21](https://github.com/yaklabco/dot/commit/c689f217e05217effe9ae01c5fb6c5f1ded35f53))
* **cli:** correct scan flag variable scope in NewDoctorCommand ([bc2074e](https://github.com/yaklabco/dot/commit/bc2074e57e1b98fb242606f5f12174992e68d3f3))
* **client:** populate packages from plan when empty in updateManifest ([3bfeb2e](https://github.com/yaklabco/dot/commit/3bfeb2e682bc3f4a0c2c369b123fbfa68ad058ee))
* **client:** properly propagate manifest errors in Doctor ([38e7d5f](https://github.com/yaklabco/dot/commit/38e7d5f62557f2ff3dce1b56d6d540122e8d4fd9))
* **cli:** handle both pointer and value operation types in renderers ([10fdb6b](https://github.com/yaklabco/dot/commit/10fdb6ba60b676bc09a4c3dfb4027d8705d4152f))
* **cli:** improve async version check and test file cleanup ([963cf85](https://github.com/yaklabco/dot/commit/963cf8563a0133b47d4059ada7ac944802a80d4f))
* **cli:** improve config format detection and help text indentation ([1581666](https://github.com/yaklabco/dot/commit/158166643bc8e7f6665e3b399240187caa78da1b))
* **cli:** improve JSON/YAML output and doctor performance ([0a29ab6](https://github.com/yaklabco/dot/commit/0a29ab6e17c7969c73bf1a2d701000f8537d0aea))
* **cli:** improve TTY detection portability and path truncation ([a52d8b4](https://github.com/yaklabco/dot/commit/a52d8b4f8ad00567e62c4a8539d217bea057d207))
* **cli:** render execution plan in dry-run mode ([39f95bc](https://github.com/yaklabco/dot/commit/39f95bc52f7b6742d4c9b92d012710ac3780d344))
* **cli:** render execution plan in dry-run mode ([9d2c0b9](https://github.com/yaklabco/dot/commit/9d2c0b99b62c1a3eeb7ee6e7a8fa480c7853b1eb))
* **cli:** replace time.Sleep race with sync.WaitGroup in ProgressTracker ([d539cfd](https://github.com/yaklabco/dot/commit/d539cfdda1df320f2215988099a8410c5bc8c88f))
* **cli:** resolve critical bugs in progress, config, and rendering ([a2fa940](https://github.com/yaklabco/dot/commit/a2fa940a48fbedd52494dd52bb9b4d4aa30eede8))
* **cli:** resolve gosec G602 array bounds warnings in pager ([2a8213b](https://github.com/yaklabco/dot/commit/2a8213be0d22b2fabce7be96daaf05ce6b2279e1))
* **cli:** respect NO_COLOR environment variable in shouldColorize ([13e722b](https://github.com/yaklabco/dot/commit/13e722bd6c9e2c66fc3411c4d1aa4a943bd069b5))
* **clone:** improve string handling safety in git SHA parsing ([4bf3b93](https://github.com/yaklabco/dot/commit/4bf3b9339c6e14fbe428e1f40b19bfcf4641b425))
* **clone:** resolve silent errors and add comprehensive logging ([c05063a](https://github.com/yaklabco/dot/commit/c05063a87560975f499f02733c5e714a2611f6a5))
* **clone:** use errors.As for wrapped error handling ([25cf928](https://github.com/yaklabco/dot/commit/25cf9285052e9ce5934ff33077226ccf50d7a2a8))
* **config:** add missing KeyDoctorCheckPermissions constant ([0b51035](https://github.com/yaklabco/dot/commit/0b5103533e8b337109ceea1e7bc70be903f0010f))
* **config:** enable CodeRabbit auto-review for all pull requests ([85eb357](https://github.com/yaklabco/dot/commit/85eb357151b537c1d9cbd2742a677d22900742ec))
* **config:** honor MarshalOptions.Indent in TOML strategy ([cfb8e0b](https://github.com/yaklabco/dot/commit/cfb8e0bf35125be679e7cb6ad4a75629708d02a8))
* **deps:** update golang.org/x/crypto to v0.43.0 to fix GO-2025-4116 ([4226dda](https://github.com/yaklabco/dot/commit/4226dda8b296a6a36eb3c917e48be7f9a50322a8))
* **doctor:** address PR review comments and improve triage robustness ([148633f](https://github.com/yaklabco/dot/commit/148633f5fbfc048e96c75378aff1b48b1ad1df39))
* **doctor:** detect and report permission errors on link targets ([bd6bd0d](https://github.com/yaklabco/dot/commit/bd6bd0d199c8b9c9264f7b03ce12b735322d0115))
* **doctor:** filter ignored links during scan ([8240491](https://github.com/yaklabco/dot/commit/8240491bcc97b4707558171d1669c73f75710f22))
* **doctor:** implement AutoIgnoreHighConfidence flag and improve input validation ([c705940](https://github.com/yaklabco/dot/commit/c705940e7ed72d8536ebc7db9392e072e62c3bec))
* **doctor:** resolve 7 major CodeRabbit issues ([13530bb](https://github.com/yaklabco/dot/commit/13530bb5cb6b5671ea6279fe03fedc3260e1ccd0))
* **doctor:** resolve linting errors and improve error handling ([83dc516](https://github.com/yaklabco/dot/commit/83dc516f3b134196886f4c96d103497d7e0b42d2))
* **domain:** make path validator tests OS-aware for Windows ([2689905](https://github.com/yaklabco/dot/commit/26899056863c2114a405e27f06c45bff22d8b4ba))
* **executor:** address code review feedback for concurrent safety and error handling ([9ca30c9](https://github.com/yaklabco/dot/commit/9ca30c9b79d221892980b2b5445a7e2f1f0132e0))
* **executor:** make Checkpoint operations map thread-safe ([6772cf6](https://github.com/yaklabco/dot/commit/6772cf6a0bc1b3bbe3376428ff363ab9612c178e))
* **hooks:** check overall project coverage to match CI ([cfb7276](https://github.com/yaklabco/dot/commit/cfb7276871c2e3d650160c883d239773fa93bd63))
* **hooks:** show linting output in pre-commit hook ([cf9d193](https://github.com/yaklabco/dot/commit/cf9d193f7cb20303cba3a132dba43f4af56df2a9))
* **hooks:** show test output in pre-commit hook ([f189550](https://github.com/yaklabco/dot/commit/f18955000d68c9f6848ad613a047a9906561d133))
* **list:** show per-package target directories from manifest ([661cf10](https://github.com/yaklabco/dot/commit/661cf10aedf0437e736832beb257b4d5ee34c6db))
* **manage:** allow empty plans and improve output consistency ([14ba5af](https://github.com/yaklabco/dot/commit/14ba5af6d1b6365918a045b996e5a5c1d578c3b9))
* **manage:** implement proper unmanage in planFullRemanage ([4dd2f54](https://github.com/yaklabco/dot/commit/4dd2f54191a4b45e98ddc47cef561539a5452dcd))
* **manifest:** add security guards and prevent hash collisions ([376ed40](https://github.com/yaklabco/dot/commit/376ed401aadc2013724759cce081775e9d7f82c8))
* **manifest:** propagate non-not-found errors in Update ([f3436cc](https://github.com/yaklabco/dot/commit/f3436ccd9517f899d96da05258d23e57582bb1a0))
* **pager:** remove blank lines left by status indicator after paging ([53c10a7](https://github.com/yaklabco/dot/commit/53c10a7d0896bd0bfdc3da4b19210c6dec86481d))
* **path:** add method forwarding to Path wrapper type ([e55b8de](https://github.com/yaklabco/dot/commit/e55b8de86b23649806788e7790b6eea1f0c595e5))
* **pipeline:** prevent shared mutation of context maps in metadata conversion ([eab124e](https://github.com/yaklabco/dot/commit/eab124e27299225e3b49c0fdf647169fdb76db90))
* **planner:** resolve directory creation dependency ordering ([0742ce9](https://github.com/yaklabco/dot/commit/0742ce91756e32fa86839eef05dbc0b6a84fa431))
* **release:** move tag to amended commit in release workflow ([1fac339](https://github.com/yaklabco/dot/commit/1fac339e2852b72141ad4fe3759971d2c26d1d8d))
* **release:** separate archive configs for Homebrew compatibility ([ca88125](https://github.com/yaklabco/dot/commit/ca88125f61d4dde5477e307277477aceffbd1a73))
* **review:** address CodeRabbit review comments for PR [#38](https://github.com/yaklabco/dot/issues/38) ([18a94d3](https://github.com/yaklabco/dot/commit/18a94d34837145d2c2206e1babb2073d753bf3b3))
* **status:** propagate non-not-found manifest errors ([38be7da](https://github.com/yaklabco/dot/commit/38be7daaa860d475a8e8db0debbeab38ea02a2df))
* **test:** add proper error handling to CLI integration tests ([e93dcdd](https://github.com/yaklabco/dot/commit/e93dcddc8d584d2d3c44192c8f11f1e7ed26f4dc))
* **test:** add Windows build constraints to Unix-specific tests ([cbe7a56](https://github.com/yaklabco/dot/commit/cbe7a56e13ba6d1dc71a514e7e54c4b770341bf4))
* **test:** add Windows compatibility to testutil symlink tests ([f26c905](https://github.com/yaklabco/dot/commit/f26c90579d9a2bedeea244bf5495b425c5805a53))
* **test:** correct comment in PlanOperationsEmpty test ([5f10f32](https://github.com/yaklabco/dot/commit/5f10f32c52d684bd36b8611c031382a674430961))
* **test:** correct mock variadic parameter handling in ports_test ([b4ecd4a](https://github.com/yaklabco/dot/commit/b4ecd4a26002383ed6a274dad29c60c837b8f063))
* **test:** improve test isolation and cross-platform compatibility ([3fb42de](https://github.com/yaklabco/dot/commit/3fb42de21342b14051b637689047f7fcaa5897a0))
* **test:** isolate XDG directories to prevent writes to source tree ([3dc6825](https://github.com/yaklabco/dot/commit/3dc682512c95cfc80ba8b6254f0d1f18d42aabea))
* **test:** make Adopt execution error test deterministic ([2ea49d3](https://github.com/yaklabco/dot/commit/2ea49d384766357d8e5c9898df74a189ae4f2bee))
* **test:** normalize working directory paths in golden tests ([5b76e66](https://github.com/yaklabco/dot/commit/5b76e66e8e94884304c78003891eeb2d2629804f))
* **test:** rename ExecutionFailure test to match actual behavior ([82b59bc](https://github.com/yaklabco/dot/commit/82b59bca6da1fa55f2b7ae6bfdfec8cd38b3a064))
* **tests:** improve permission conflict test robustness ([bc0660c](https://github.com/yaklabco/dot/commit/bc0660cf065954595c4e4aec52b78161488937e5))
* **test:** skip file mode test on Windows ([fe934b3](https://github.com/yaklabco/dot/commit/fe934b3a399b84ad00fc4a829e71b9c15256ff9c))
* **tests:** skip permission test in CI environments ([2a7f5a2](https://github.com/yaklabco/dot/commit/2a7f5a2852812381a708dad951cafac1c6adb149))
* **test:** update tests for new conflict detection behavior ([05d98f5](https://github.com/yaklabco/dot/commit/05d98f597ab4821784a621be9a5b68c7b8d2f8be))
* **test:** use cmd.OutOrStdout for adopt command output ([3bd2196](https://github.com/yaklabco/dot/commit/3bd2196ad769c1519fdd25171cdb0b0d6bf2072f))
* **ui:** add newline after table output for better terminal spacing ([7c8ad54](https://github.com/yaklabco/dot/commit/7c8ad54285f499a5325dfd4e62f6040122eb1c96))
* **unmanage:** improve corrupted structure handling ([0030034](https://github.com/yaklabco/dot/commit/00300343ecafacff7b5917d4277db35df05c4603))
* **unmanage:** use filepath.Join for cross-platform path handling ([0b01663](https://github.com/yaklabco/dot/commit/0b0166323ef0b1c868b25d1ed00c9675e4972d2a))
* **updater:** address critical bugs and security issues ([959f225](https://github.com/yaklabco/dot/commit/959f2254288e42b87ee0d33e379ddc89a195d613))
* **updater:** correct notification box alignment and version truncation ([8e1184e](https://github.com/yaklabco/dot/commit/8e1184edd757b7bf16f7c27f513ae3adbe9c4cf4))
* **updater:** handle development versions and improve notification UX ([3e315c4](https://github.com/yaklabco/dot/commit/3e315c44bb10db313bc1f876906913419dddd817))
* **vuln:** exclude GO-2024-3295 from vulnerability checks ([41e5625](https://github.com/yaklabco/dot/commit/41e56251dfc58a799d4de431f027f3122c1ff842))
* **vuln:** extract findings only, not all OSV database entries ([ba7c020](https://github.com/yaklabco/dot/commit/ba7c020abf3df87fbc062fd1485446cb04e0a931))


### Performance Improvements

* **doctor:** optimize scan performance with parallel execution and smart filtering ([a4cfd79](https://github.com/yaklabco/dot/commit/a4cfd79c9b69433b6c987287aa96950b7b45da9a))
* **pipeline:** use targeted path scanning instead of recursive directory scan ([be6f8a8](https://github.com/yaklabco/dot/commit/be6f8a8b6ab0a23c05177a3410a612444b432380))


### Code Refactoring

* **adopt:** implement flat package structure with consistent dot-prefix ([0b8b2e7](https://github.com/yaklabco/dot/commit/0b8b2e706506710a1d968350e9536ea1a145dc82))
* **adopt:** preserve leading dots in package naming ([3a0075b](https://github.com/yaklabco/dot/commit/3a0075bb347fd82541a78a29a997712957e4596a))
* **adopt:** update Adopt and PlanAdopt methods to use files-first signature ([29b7d5b](https://github.com/yaklabco/dot/commit/29b7d5bcdb4690e41198289f28be0a6a2b304052))
* **api:** extract orphan scan logic to reduce complexity ([3d18f77](https://github.com/yaklabco/dot/commit/3d18f77386d875c198354089b3822995da69987a))
* **api:** reduce cyclomatic complexity in PlanRemanage ([fa95354](https://github.com/yaklabco/dot/commit/fa9535495f2a965445f6c6f2a619e4eec962dad7))
* **api:** replace Client interface with concrete struct ([0eae844](https://github.com/yaklabco/dot/commit/0eae8443892c481a08f6feeedbd32953635285af))
* **architecture:** implement domain separation and Client struct conversion ([16edc7f](https://github.com/yaklabco/dot/commit/16edc7f5ba9511d0bca58027228da2b00ce751b2))
* **bootstrap:** remove unused makeSet helper function ([d210c9c](https://github.com/yaklabco/dot/commit/d210c9caabd42961ba62faf52cbbe7e2a7e72ae5))
* **cli:** add default case and eliminate type assertion duplication ([c227880](https://github.com/yaklabco/dot/commit/c227880ad87a0acf2fe28a4e0b9c9d87cb231052))
* **cli:** address code review nitpicks for improved code quality ([7664bf9](https://github.com/yaklabco/dot/commit/7664bf9b27278725a44bbebba2df89c103540bed))
* **cli:** consolidate color system and standardize output formatting ([873ff0e](https://github.com/yaklabco/dot/commit/873ff0eb236172791718886512def16c7e65c086))
* **cli:** consolidate color system and standardize output formatting ([eee86ab](https://github.com/yaklabco/dot/commit/eee86ab3dd374342ac71a80c8afef4d24d822b42))
* **cli:** enhance unmanage output and add -y flag ([dcf452b](https://github.com/yaklabco/dot/commit/dcf452b82038c18d2fb23dcaa0fcbf9f7bde8f7d))
* **cli:** improve config loading, progress tracker, and command UX ([610dfe8](https://github.com/yaklabco/dot/commit/610dfe8a03ebed009028bd73c5db39f42267ec14))
* **cli:** improve success messages with proper pluralization ([efb49ed](https://github.com/yaklabco/dot/commit/efb49ed96a6f36410b077b5767a9b63994a3864a))
* **cli:** make cleanup grace period a constant ([b28df76](https://github.com/yaklabco/dot/commit/b28df76fd3cbd2873a5c6ccbda349f9b61cbd7db))
* **cli:** migrate from go-pretty to lipgloss v1.1.0 ([f7244c3](https://github.com/yaklabco/dot/commit/f7244c3e9e4080a40aae4fceb5a39d1b4bf0bc4c))
* **cli:** reduce cyclomatic complexity in table renderer ([2ab45e2](https://github.com/yaklabco/dot/commit/2ab45e24e786f21b823d9e3897f9b903c9783899))
* **cli:** remove redundant goroutine wrapper ([e61309e](https://github.com/yaklabco/dot/commit/e61309e92ba3706dc03fa029b0109be1b2bf4f6f))
* **config:** migrate writer to use strategy pattern ([e1e1878](https://github.com/yaklabco/dot/commit/e1e1878e44e9c174c46dd9bfe3aa15079ce4272d))
* **config:** simplify repository configuration loading ([ceaa12e](https://github.com/yaklabco/dot/commit/ceaa12ecc7a2919dcff56c9040a18b674e6f7968))
* **config:** use permission constants ([c794619](https://github.com/yaklabco/dot/commit/c7946194fa9f5ca7cfcd4b23f000a37d37348e6a))
* **doctor:** decouple checks and fix import cycles ([dfe8d32](https://github.com/yaklabco/dot/commit/dfe8d32a6dd736629baf4f61a2de08c4b70ce19a))
* **doctor:** enhance orphan scan with worker context and result collection ([4394afd](https://github.com/yaklabco/dot/commit/4394afd3da90cf6eecf272d34ce9c88559a5cce3))
* **doctor:** remove dead code and respect dry-run in category ignore ([d938541](https://github.com/yaklabco/dot/commit/d9385410f2bfc77214b174ffdc13c28535971ccc))
* **domain:** clean up temporary migration scripts ([7fd94a2](https://github.com/yaklabco/dot/commit/7fd94a25b13d480b7a40c3e3f59201c50da4b544))
* **domain:** complete internal package migration and simplify pkg/dot ([fe638e3](https://github.com/yaklabco/dot/commit/fe638e308003274112e817716bfe0605dfe2784e))
* **domain:** create internal/domain package structure ([4295770](https://github.com/yaklabco/dot/commit/4295770b8579542e60db7956f7233b99356303a0))
* **domain:** format code and fix linter issues ([1cd949e](https://github.com/yaklabco/dot/commit/1cd949e7e79ed85f57ca76e96afd64e0ffdcec82))
* **domain:** improve TraversalFreeValidator implementation ([b7c81c6](https://github.com/yaklabco/dot/commit/b7c81c61943021c6b415180a16ce6bf02ff3166f))
* **domain:** move all domain types to internal/domain ([a1533d0](https://github.com/yaklabco/dot/commit/a1533d0b3ddb24120a43fc7e0d2a3af2c4495c97))
* **domain:** move MustParsePath to testing.go ([834127f](https://github.com/yaklabco/dot/commit/834127fc3543e437043c7244e665f91d1005b573))
* **domain:** move Path and errors types to internal/domain ([f058689](https://github.com/yaklabco/dot/commit/f058689198900abf6148b65e2abbf7c61cf4995b))
* **domain:** move Result monad to internal/domain ([a4f6009](https://github.com/yaklabco/dot/commit/a4f600983571879836fbbbf2efb4a3fef05f4427))
* **domain:** update all internal package imports to use internal/domain ([c3b057f](https://github.com/yaklabco/dot/commit/c3b057f44d366918ac4d934dc520f38679ec9cb8))
* **domain:** use TargetPath for operation targets ([929f47c](https://github.com/yaklabco/dot/commit/929f47c275cfc3c50167804b2800278827756c3c))
* **domain:** use validators in path constructors ([31548dd](https://github.com/yaklabco/dot/commit/31548dd8232b1199eda1d7dfd2398d250b229fed))
* **dotprefix:** work in progress on dot prefix refactoring ([12a1190](https://github.com/yaklabco/dot/commit/12a11908b7c3a4228df098eee0a5ed70b9c37b4f))
* **hooks:** eliminate duplicate test run in pre-commit ([c3a47d8](https://github.com/yaklabco/dot/commit/c3a47d8c5f15b221b92eaa51d445ca6dfa3dd504))
* **path:** remove Path generic wrapper to eliminate code quality issue ([79910c6](https://github.com/yaklabco/dot/commit/79910c678384bd669ac7c53791576185188f7803))
* **pipeline:** improve test quality and organization ([fdf75d0](https://github.com/yaklabco/dot/commit/fdf75d07be7a81feb5bbaa96184ac7219df4cc41))
* **pipeline:** use safe unwrap pattern in path construction tests ([2756a41](https://github.com/yaklabco/dot/commit/2756a413304823248c2e4ba31411fe4bb5e8648c))
* **pkg:** code quality improvements ([fd45d7a](https://github.com/yaklabco/dot/commit/fd45d7ab8aab187a536c8f3a108173b908461101))
* **pkg:** convert Client to facade pattern ([5edd49f](https://github.com/yaklabco/dot/commit/5edd49fed856bac7aae10f586cc898248fa1c44f))
* **pkg:** extract helper methods in DoctorService ([3069a10](https://github.com/yaklabco/dot/commit/3069a106435a6b2b11d26079f19c79de9431ba6f))
* **pkg:** replace MustParsePath with error handling in production code ([fd7f876](https://github.com/yaklabco/dot/commit/fd7f876de7982172bb9451f926007ac24eed2dbf))
* **pkg:** simplify DoctorWithScan method ([028e2b4](https://github.com/yaklabco/dot/commit/028e2b4e6073c7d210280002faf126555e918a23))
* **pkg:** simplify scanForOrphanedLinks method ([9824eea](https://github.com/yaklabco/dot/commit/9824eea0b1f9e9ef491c87c0bd8138ab8ecfd154))
* **quality:** improve error handling documentation and panic messages ([31a483e](https://github.com/yaklabco/dot/commit/31a483e955729872acf7c833aa6370b69c649cb6))
* **terminology:** complete symlink removal from test fixtures ([8656740](https://github.com/yaklabco/dot/commit/86567409b0902d247c3c308074e05819e0ec59fe))
* **terminology:** rename symlink-prefixed variables to package/manage ([57703b5](https://github.com/yaklabco/dot/commit/57703b577928a0a453bc0815a46ef3b5af08617c))
* **terminology:** replace symlink with package directory terminology ([d75a03a](https://github.com/yaklabco/dot/commit/d75a03a5e93051b414fdccccbce35daf71733db4))
* **terminology:** update suggestion text from unmanage to unmanage ([09a9eec](https://github.com/yaklabco/dot/commit/09a9eecd4fc48162e0af6d312f8ce4fb07ec0ca1))
* **test:** improve benchmark tests with proper error handling ([63c1f94](https://github.com/yaklabco/dot/commit/63c1f94a2fc5187ec1b5a41732db03c40c0b0cbb))
* **updater:** optimize color detection and ANSI stripping ([f7967bd](https://github.com/yaklabco/dot/commit/f7967bdbd697425b68435b75b95512811c279c54))


### Tests

* **adapters:** add comprehensive MemFS tests to achieve 80%+ coverage ([602560a](https://github.com/yaklabco/dot/commit/602560a3dabe9736aaccef1078ecea1ed2076bae))
* **adapters:** replace network-dependent tests with hermetic fixtures ([1907fbe](https://github.com/yaklabco/dot/commit/1907fbe748e2db7fafe026e88d22211015a9243a))
* add comprehensive test coverage ([28d245f](https://github.com/yaklabco/dot/commit/28d245fb4ca17ccdc8f4422f6fcc1fb2fffbf337))
* add fuzz tests for config, domain, and ignore packages ([762db0c](https://github.com/yaklabco/dot/commit/762db0cd667fa2fc57fe0020fe9108485c3abb11))
* **adopt:** add regression tests for adoption bugs ([ec2cf99](https://github.com/yaklabco/dot/commit/ec2cf99122ed5f3e155e835b79d7c927930be233))
* **api,dot:** add helper and operation ID tests ([86e7271](https://github.com/yaklabco/dot/commit/86e72716d9af890151c7b2fd17db648f6ff5dbd9))
* **api:** add comprehensive test coverage for all API methods ([0aadfac](https://github.com/yaklabco/dot/commit/0aadfacb4a7952b71973af873b4eef91f123fba1))
* **api:** add manifest helper tests and document remediation ([4a123b4](https://github.com/yaklabco/dot/commit/4a123b41801a2ce5661401729c72dc7fb8accb8d))
* **backup:** add comprehensive integration tests for backup workflow ([8992f28](https://github.com/yaklabco/dot/commit/8992f28302a4928de7a63e213073a19f537075a6))
* **bootstrap:** add manifest filtering verification assertions ([f145318](https://github.com/yaklabco/dot/commit/f1453184b4cac375a77617adbce19d7a6c75899b))
* **cli:** add comprehensive main package test coverage ([a09c38f](https://github.com/yaklabco/dot/commit/a09c38f391ad293fc6434040d29786de40b25484))
* **cli:** add golden file testing framework ([70041a5](https://github.com/yaklabco/dot/commit/70041a55421804bb41a31b6a94bb2369f8450757))
* **cli:** add golden tests for adopt and manage commands ([b3999a6](https://github.com/yaklabco/dot/commit/b3999a662d71cbe572dfc930c9e0c46b9c4a0ab7))
* **cli:** add signal handling integration tests ([832fd74](https://github.com/yaklabco/dot/commit/832fd74d237c4e2ea9d86b2c4a392313f590e040))
* **cli:** complete runtime error test assertions ([ba4556e](https://github.com/yaklabco/dot/commit/ba4556e6fd1908aa4bf69b8e3657ee3b2d92673d))
* **client:** add comprehensive tests for pkg/dot Client struct ([96fded7](https://github.com/yaklabco/dot/commit/96fded786bb97ec23ef425f0409bcc3bcc8670b6))
* **client:** add edge case tests for coverage buffer ([710a1e1](https://github.com/yaklabco/dot/commit/710a1e198b0650a56b47d3cda435851467d76a23))
* **client:** add exhaustive tests for increased coverage margin ([19183f5](https://github.com/yaklabco/dot/commit/19183f51483c59d9417c552041f42df30bb4abf8))
* **cli:** fix help text assertion after symlink removal ([a4f113a](https://github.com/yaklabco/dot/commit/a4f113ae88b188807dd169ea4670a96d138f3608))
* **cli:** fix verification test expectations ([a026414](https://github.com/yaklabco/dot/commit/a026414bed20bedcddef056798f052f916a1f750))
* **cli:** improve signal handling test isolation ([ce62014](https://github.com/yaklabco/dot/commit/ce62014a65f539ea320a394f61bb0ee45c7f6654))
* **cli:** increase cmd/dot test coverage to 88.6% ([5e49359](https://github.com/yaklabco/dot/commit/5e4935960683fd04a1583b254bbbb1e7238e7c17))
* **clone:** add coverage for auth method name formatting ([25e1140](https://github.com/yaklabco/dot/commit/25e11406c232fc836c9739b6a15165d4fe8b5ea3))
* **cmd:** add basic command constructor tests ([aa06c3a](https://github.com/yaklabco/dot/commit/aa06c3a27df1fcf14dd2df3712ad16da511bab39))
* **config:** add aggressive coverage boost tests ([fdfb8e3](https://github.com/yaklabco/dot/commit/fdfb8e324704839cdd68b9eeb00f09e2d4df6355))
* **config:** add comprehensive loader and precedence tests ([62cabe1](https://github.com/yaklabco/dot/commit/62cabe173e35fa194e04c814cff2fc8843d5b3c1))
* **config:** add configuration key constant tests ([b271faa](https://github.com/yaklabco/dot/commit/b271faafcd3d9e20ed52f9b95ddec43ca4c59d4a))
* **config:** add default value constant tests ([ddce2ca](https://github.com/yaklabco/dot/commit/ddce2ca4d994adba258b64a9a71d4d5a3ad697c2))
* **config:** add marshal strategy interface tests ([af407d5](https://github.com/yaklabco/dot/commit/af407d5c7276029797f59f550d45156e8336e405))
* **config:** add validation edge case tests ([fbe593b](https://github.com/yaklabco/dot/commit/fbe593b54b5d5ebe574c95203b175019c025553e))
* **doctor:** improve context cancellation test ([94efc24](https://github.com/yaklabco/dot/commit/94efc2445ae278bc1fb1017cfc44073366070b28))
* **domain:** add error helper tests ([457b0e1](https://github.com/yaklabco/dot/commit/457b0e14593f0c98e56e2dc2e1658f10abc78e52))
* **domain:** add path validator tests ([ad682cb](https://github.com/yaklabco/dot/commit/ad682cb0f027ef518a1da4f890275c0f3531a25f))
* **domain:** add permission constant tests ([2822e1f](https://github.com/yaklabco/dot/commit/2822e1f0bbc29224f27ffe041fd79834a9ecaa94))
* **domain:** add Result unwrap helper tests ([e8a0208](https://github.com/yaklabco/dot/commit/e8a02081ecaff45e44b883fcf7dd7834a0c77ff1))
* **dot:** add comprehensive error and operation tests ([d13573a](https://github.com/yaklabco/dot/commit/d13573a96befbbe5ad57397ee9c1e905a517fc04))
* **dot:** add tests for BootstrapService ([83cfc88](https://github.com/yaklabco/dot/commit/83cfc88cadd3b4e30415cbfb17b161675863caa2))
* **executor:** add comprehensive tests to exceed 80% coverage threshold ([810d5a8](https://github.com/yaklabco/dot/commit/810d5a8707167a61e880f656129def5480ce5381))
* **integration:** add clone feature tests and fixtures ([54228f2](https://github.com/yaklabco/dot/commit/54228f2649286465df3f4af84650cd066f6921fe))
* **integration:** implement comprehensive integration test infrastructure ([4521328](https://github.com/yaklabco/dot/commit/4521328f915ef2d5e61c8d175a322d75b56b20f4))
* **integration:** implement comprehensive integration testing ([3304c94](https://github.com/yaklabco/dot/commit/3304c946aa6c91a53a65d164d3b3e80ea10d6bdb))
* **integration:** implement integration test categories ([adac35b](https://github.com/yaklabco/dot/commit/adac35b72647dfd60314cad8042454a237a6f79c))
* **pipeline:** add coverage for operation mapping and state scanner ([7cc780b](https://github.com/yaklabco/dot/commit/7cc780b8b6d86ba1035dd40961e04537c49b1f58))
* **quality:** implement code review remediation ([e6c1282](https://github.com/yaklabco/dot/commit/e6c1282c25f65a5c8e19c75ffd0f68205393dacf))
* **quality:** increase coverage to 80.2% and fix quality issues ([ef6a9ad](https://github.com/yaklabco/dot/commit/ef6a9ad1d6ef4ba15ba5a8e797dd4e833b73e8a5))
* **scanner:** add benchmarks for package scanning performance ([56390f1](https://github.com/yaklabco/dot/commit/56390f18b88ecc3c34b2446fc2e0657c2e9f8a10))
* **status:** add health checking tests ([00443e6](https://github.com/yaklabco/dot/commit/00443e60156ad7a12d4501f4ab3c29cb4cd7861d))
* **terminal:** add comprehensive tests for terminal detection ([95bdc13](https://github.com/yaklabco/dot/commit/95bdc13412c9bba688ce00032d6d7b7ca48d7e9e))


### Build System

* **changelog:** implement automated changelog generation with git-chglog ([93e5b24](https://github.com/yaklabco/dot/commit/93e5b24e5b52761b638d741c77637cbe7a6c4361))
* **deps:** add go-git and golang.org/x/term dependencies ([69d33e5](https://github.com/yaklabco/dot/commit/69d33e5dffbfa53e6008e1de33b5c4749266cf0a))
* **deps:** update golang.org/x/crypto to v0.45.0 ([3b07074](https://github.com/yaklabco/dot/commit/3b070741738b1ef6760073249e2343828f53bf0b))
* **distribution:** add Homebrew tap with automated releases ([ab63e64](https://github.com/yaklabco/dot/commit/ab63e649dc7fc3edc2e55ecad341321d4285a37d))
* **make:** add buildvcs flag for reproducible builds ([d027da5](https://github.com/yaklabco/dot/commit/d027da557b3b6d4da4b88e1720160ac3e05d9986))
* **makefile:** add coverage threshold validation to check target ([703ff2c](https://github.com/yaklabco/dot/commit/703ff2c454ad85876c33472a52eddf03706682d1))
* **makefile:** add uninstall target to remove installed binary ([53b9b3b](https://github.com/yaklabco/dot/commit/53b9b3bde889cdbfcb145a2d342b6bf31164091c))
* **release:** add Homebrew tap integration ([3df890f](https://github.com/yaklabco/dot/commit/3df890fc16580bb0b55b274c55f4fa099aede7c1))


### Continuous Integration

* add vulnerability checking with govulncheck ([6f55c20](https://github.com/yaklabco/dot/commit/6f55c2028996a74720faa56a205fd645d436d601))
* **release:** implement Release Please automation ([698e1b5](https://github.com/yaklabco/dot/commit/698e1b57dc5b2e7c85f3cdc3b2730928e43a62fd))
* **release:** install golangci-lint before running linters ([42ff560](https://github.com/yaklabco/dot/commit/42ff560106cfc7340a7d8a6621b7a5b9c9f092e8))
* **release:** use GORELEASER_TOKEN for tap updates ([7f819d1](https://github.com/yaklabco/dot/commit/7f819d1e0e74ec860938c07afc91634ffd341360))
* **vuln:** use JSON parsing for reliable vulnerability exclusion ([e03baef](https://github.com/yaklabco/dot/commit/e03baefda9589707faacdb674ec0ee229939e636))

## [Unreleased]


<a name="v0.5.0"></a>
## [v0.5.0] - 2025-11-02
### Chore
- **git:** add test artifacts to .gitignore

### Ci
- **vuln:** use JSON parsing for reliable vulnerability exclusion

### Docs
- **changelog:** update for v0.5.0 release

### Fix
- **test:** isolate XDG directories to prevent writes to source tree
- **test:** normalize working directory paths in golden tests
- **test:** use cmd.OutOrStdout for adopt command output
- **vuln:** extract findings only, not all OSV database entries


<a name="v0.4.4"></a>
## [v0.4.4] - 2025-11-02
### Chore
- fix the bootstrap config
- fix the bootstrap config
- update IDE exclusions in .gitignore
- **build:** add fuzz and bench targets to Makefile

### Ci
- add vulnerability checking with govulncheck

### Docs
- add security policy with vulnerability disclosure
- remove stale and broken links from documentation
- **changelog:** update for v0.4.4 release
- **cli:** improve clone help text scannability
- **readme:** update adopt command documentation

### Feat
- **adapters:** integrate GitHub CLI authentication
- **api:** integrate backup system with manage service
- **cli:** add batch mode and async version check
- **cli:** add color helper for dynamic color detection
- **cli:** add profiling and diagnostics support
- **cli:** add command aliases for config
- **cli:** add pluralization helpers for output formatting
- **cli:** improve package selection UI with colors and layout
- **cli:** add interactive prompt utilities
- **clone:** derive directory from repository name like git clone
- **config:** add network configuration support
- **config:** add backup and overwrite configuration options
- **config:** add validation for network timeout fields
- **domain:** add FileDelete operation and fix FileBackup permissions
- **manifest:** add backup tracking to package metadata
- **pipeline:** add current state scanner for conflict detection
- **planner:** implement backup and overwrite conflict policies
- **retry:** add exponential backoff retry utility
- **updater:** add security validation for package managers
- **updater:** improve version checking with better error handling

### Fix
- **cli:** improve async version check and test file cleanup
- **cli:** resolve gosec G602 array bounds warnings in pager
- **test:** update tests for new conflict detection behavior
- **vuln:** exclude GO-2024-3295 from vulnerability checks

### Refactor
- **cli:** make cleanup grace period a constant
- **cli:** remove redundant goroutine wrapper
- **cli:** enhance unmanage output and add -y flag
- **cli:** improve success messages with proper pluralization

### Test
- add fuzz tests for config, domain, and ignore packages
- **backup:** add comprehensive integration tests for backup workflow
- **cli:** fix verification test expectations
- **cli:** improve signal handling test isolation
- **cli:** add golden tests for adopt and manage commands
- **cli:** add golden file testing framework
- **cli:** add comprehensive main package test coverage
- **cli:** add signal handling integration tests
- **clone:** add coverage for auth method name formatting
- **pipeline:** add coverage for operation mapping and state scanner
- **scanner:** add benchmarks for package scanning performance

### Pull Requests
- Merge pull request [#36](https://github.com/yaklabco/dot/issues/36) from jamesainslie/feature-improve-testing
- Merge pull request [#35](https://github.com/yaklabco/dot/issues/35) from jamesainslie/feature-improve-ux
- Merge pull request [#34](https://github.com/yaklabco/dot/issues/34) from jamesainslie/feature-use-go-gh-sdk-for-auth

### BREAKING CHANGE

```
Remove automatic glob mode detection to eliminate ambiguous behavior. Users must now provide explicit package names when adopting multiple files.

Changes:
- Remove fileExists(), deriveCommonPackageName(), and commonPrefix()
- Simplify logic: single file = auto-naming, multiple = explicit
- Update help text with section headers for clarity
- Remove tests for deleted helper functions
- Update success message format

Before:
  dot adopt .git*  # Auto-detected package name from files

After:
  dot adopt git .git*  # Explicit package name required

This ensures predictable behavior and clearer package organization.
```



<a name="v0.4.3"></a>
## [v0.4.3] - 2025-10-13
### Docs
- **changelog:** update for v0.4.3 release
- **user:** add comprehensive upgrade and version management documentation

### Feat
- **cli:** integrate startup version check into root command
- **cli:** implement dot upgrade command with comprehensive testing
- **config:** add update configuration for package manager and version checking
- **updater:** add colorized output to update notification
- **updater:** implement startup version checking system
- **updater:** add version checker and package manager services

### Fix
- **updater:** address critical bugs and security issues
- **updater:** correct notification box alignment and version truncation

### Refactor
- **updater:** optimize color detection and ANSI stripping

### Test
- **quality:** increase coverage to 80.2% and fix quality issues

### Pull Requests
- Merge pull request [#33](https://github.com/yaklabco/dot/issues/33) from jamesainslie/feature-upgrade-command


<a name="v0.4.2"></a>
## [v0.4.2] - 2025-10-12
### Docs
- **changelog:** update for v0.4.2 release
- **readme:** add clone and bootstrap commands to documentation

### Feat
- **cli:** refactor table rendering with go-pretty for professional UX
- **cli:** add progress tracking using go-pretty
- **cli:** add list and text formatting utilities
- **cli:** add table infrastructure using go-pretty
- **cli:** add muted colorization to doctor and unmanage commands
- **config:** wire up table_style configuration to all commands
- **doctor:** enable scoped orphan scanning by default and detect broken unmanaged links
- **pager:** add keyboard controls for interactive pagination
- **ui:** add automatic pagination to dot doctor output
- **ui:** add configuration toggle between modern and legacy table styles
- **unmanage:** add --all flag to unmanage all packages at once

### Fix
- **cli:** replace time.Sleep race with sync.WaitGroup in ProgressTracker
- **clone:** resolve silent errors and add comprehensive logging
- **pager:** remove blank lines left by status indicator after paging
- **ui:** add newline after table output for better terminal spacing
- **unmanage:** use filepath.Join for cross-platform path handling

### Perf
- **doctor:** optimize scan performance with parallel execution and smart filtering

### Refactor
- **cli:** improve config loading, progress tracker, and command UX
- **cli:** migrate from go-pretty to lipgloss v1.1.0
- **config:** simplify repository configuration loading
- **doctor:** enhance orphan scan with worker context and result collection

### Test
- **terminal:** add comprehensive tests for terminal detection

### Yak
- **ci:** have a simple - pass fail on test coverage

### Pull Requests
- Merge pull request [#32](https://github.com/yaklabco/dot/issues/32) from jamesainslie/fix-some-missed-bugs
- Merge pull request [#31](https://github.com/yaklabco/dot/issues/31) from jamesainslie/fix-silent-errors-when-cloning


<a name="v0.4.1"></a>
## [v0.4.1] - 2025-10-10
### Docs
- **changelog:** update for v0.4.1 release
- **clone:** add bootstrap subcommand documentation

### Feat
- **bootstrap:** implement bootstrap config generator
- **clone:** implement clone bootstrap subcommand
- **dot:** add bootstrap generation to Client facade

### Fix
- **bootstrap:** use installed parameter in YAML comments
- **bootstrap:** implement manifest parsing for from-manifest flag
- **bootstrap:** correct documentation URL in generated config header

### Refactor
- **bootstrap:** remove unused makeSet helper function

### Test
- **bootstrap:** add manifest filtering verification assertions
- **dot:** add tests for BootstrapService

### Pull Requests
- Merge pull request [#30](https://github.com/yaklabco/dot/issues/30) from jamesainslie/feature-bootstrap-generation


<a name="v0.4.0"></a>
## [v0.4.0] - 2025-10-10
### Build
- **deps:** add go-git and golang.org/x/term dependencies

### Change
- **docs:** logo > transparent background

### Docs
- **changelog:** update for v0.4.0 release
- **changelog:** wrap BREAKING CHANGE content in code blocks
- **changelog:** regenerate from cleaned commit history
- **readme:** add clone command to Quick Start
- **user:** add clone command and bootstrap config documentation

### Feat
- **bootstrap:** add configuration schema and loader
- **cli:** implement dot clone command
- **cli:** add interactive package selector and terminal detection
- **client:** integrate CloneService into Client facade
- **clone:** improve branch detection and profile filtering
- **clone:** implement CloneService orchestrator
- **errors:** add clone-specific error types
- **git:** add git cloning with authentication support
- **manifest:** add repository tracking support

### Fix
- **adapters:** use correct HTTP Basic Auth format for tokens
- **clone:** improve string handling safety in git SHA parsing
- **clone:** use errors.As for wrapped error handling

### Test
- **adapters:** replace network-dependent tests with hermetic fixtures
- **integration:** add clone feature tests and fixtures

### Pull Requests
- Merge pull request [#29](https://github.com/yaklabco/dot/issues/29) from jamesainslie/feature-clone-command


<a name="v0.3.1"></a>
## [v0.3.1] - 2025-10-09
### Build
- **makefile:** add uninstall target to remove installed binary

### Docs
- **changelog:** update for v0.3.1 release
- **readme:** update current version to v0.3.1


<a name="v0.3.0"></a>
## [v0.3.0] - 2025-10-09
### Build
- **makefile:** add coverage threshold validation to check target

### Docs
- update documentation for v0.3 flat structure
- **adopt:** add glob expansion examples to documentation
- **changelog:** update for v0.3.0 release
- **changelog:** fix BREAKING CHANGE formatting for v0.2.0
- **pkg:** update implementation status and test comments

### Feat
- **adopt:** add auto-naming and glob expansion modes
- **cli:** add comprehensive tab-completion and unmanage restoration

### Fix
- **ci:** remove path to golangci-lint
- **ci:** remove path to golangci-lint
- **planner:** resolve directory creation dependency ordering
- **tests:** skip permission test in CI environments
- **tests:** improve permission conflict test robustness

### Refactor
- **adopt:** implement flat package structure with consistent dot-prefix
- **adopt:** preserve leading dots in package naming
- **dotprefix:** work in progress on dot prefix refactoring
- **pkg:** replace MustParsePath with error handling in production code

### Pull Requests
- Merge pull request [#28](https://github.com/yaklabco/dot/issues/28) from jamesainslie/refactor-dotprefix
- Merge pull request [#27](https://github.com/yaklabco/dot/issues/27) from jamesainslie/fix-changelog-v0.2.0-formatting
- Merge pull request [#26](https://github.com/yaklabco/dot/issues/26) from jamesainslie/pre-release-niggles

### BREAKING CHANGE

```
Adopt now creates flat package structure

Change adopt behavior to store directory contents at package root with
consistent 'dot-' prefix application.

Before:
~/dotfiles/ssh/dot-ssh/config → ~/.ssh

After:
~/dotfiles/dot-ssh/config → ~/.ssh

Changes:
- Package names preserve leading dots: .ssh → dot-ssh
- Directory contents stored at package root (flat structure)
- Apply dotfile translation to each file/directory
- Symlinks point to package root (not nested subdirectory)

Implementation:
- Add createDirectoryAdoptOperations for directory handling
- Add collectDirectoryFiles for recursive collection
- Add translatePathComponents for per-component translation
- Update unmanage restoration to handle both structures

Testing:
- All existing tests updated for new structure
- New tests for flat structure and nested dotfiles
- Regression tests preserve backward compatibility checks
- 80%+ coverage maintained

Refs: docs/planning/dot-prefix-refactoring-plan.md
```



<a name="v0.2.0"></a>
## [v0.2.0] - 2025-10-08
### Docs
- **changelog:** update for v0.2.0 release
- **developer:** add a mascot..because gopher
- **packages:** update user documentation for package name mapping

### Feat
- **cli:** display usage help on invalid flags and arguments
- **packages:** enable package name to target directory mapping

### Test
- **cli:** complete runtime error test assertions

### BREAKING CHANGE

```
Package structure requirements changed.

Before (v0.1.x):
dot-gnupg/
├── common.conf → ~/common.conf
└── public-keys.d/ → ~/public-keys.d/

After (v0.2.0):
dot-gnupg/
├── common.conf → ~/.gnupg/common.conf
└── public-keys.d/ → ~/.gnupg/public-keys.d/

Migration: Restructure packages to remove redundant nesting, or
opt-out by setting dotfile.package_name_mapping: false in config.

Rationale: Project is pre-1.0 (v0.1.1), establishing intuitive
design before API stability commitment in 1.0.0 release.
```



<a name="v0.1.1"></a>
## [v0.1.1] - 2025-10-08
### Chore
- remove documentation from source control
- remove reference docs from source control
- remove migration docs from source control
- add planning docs to gitignore
- remove planning and archive docs from source control
- **ci:** expunge emojis
- **ci:** expunge emojis
- **ci:** ignore some planning docs
- **ci:** expunge emojis
- **docs:** remove unwanted docs
- **hooks:** add pre-commit hook for test coverage enforcement

### Docs
- document refactoring
- add executive summary of completion
- update progress doc to reflect completion
- document Core completion
- add complete summary
- **architecture:** add comprehensive architecture documentation
- **architecture:** update for service-based architecture
- **changelog:** update README.md
- **changelog:** update for v0.1.1 release
- **changelog:** update for v0.1.1 release
- **changelog:** update for v0.1.1 release
- **developer:** add mermaid diagrams and comprehensive testing documentation
- **index:** update documentation index to reflect current structure
- **navigation:** add root README links to all child documentation
- **planning:** add progress checkpoint
- **planning:** add code quality improvements plan
- **readme:** fix broken documentation links
- **test:** update benchmark template to use testing.TB pattern

### Feat
- **config:** implement TOML marshal strategy
- **config:** implement JSON marshal strategy
- **config:** implement YAML marshal strategy
- **domain:** add chainable Result methods
- **pkg:** extract DoctorService from Client
- **pkg:** extract AdoptService from Client
- **pkg:** extract StatusService from Client
- **pkg:** extract UnmanageService from Client
- **pkg:** extract ManageService from Client
- **pkg:** extract ManifestService from Client
- **release:** automate release workflow with integrated changelog generation

### Fix
- **client:** properly propagate manifest errors in Doctor
- **client:** populate packages from plan when empty in updateManifest
- **config:** add missing KeyDoctorCheckPermissions constant
- **config:** honor MarshalOptions.Indent in TOML strategy
- **doctor:** detect and report permission errors on link targets
- **domain:** make path validator tests OS-aware for Windows
- **hooks:** check overall project coverage to match CI
- **hooks:** show linting output in pre-commit hook
- **hooks:** show test output in pre-commit hook
- **manage:** implement proper unmanage in planFullRemanage
- **manifest:** propagate non-not-found errors in Update
- **path:** add method forwarding to Path wrapper type
- **release:** move tag to amended commit in release workflow
- **status:** propagate non-not-found manifest errors
- **test:** strengthen PackageOperations assertion in exhaustive test
- **test:** rename ExecutionFailure test to match actual behavior
- **test:** correct comment in PlanOperationsEmpty test
- **test:** add Windows build constraints to Unix-specific tests
- **test:** add proper error handling to CLI integration tests
- **test:** add Windows compatibility to testutil symlink tests
- **test:** skip file mode test on Windows
- **test:** correct mock variadic parameter handling in ports_test

### Refactor
- **api:** replace Client interface with concrete struct
- **config:** migrate writer to use strategy pattern
- **config:** use permission constants
- **domain:** clean up temporary migration scripts
- **domain:** complete internal package migration and simplify pkg/dot
- **domain:** create internal/domain package structure
- **domain:** move Result monad to internal/domain
- **domain:** move Path and errors types to internal/domain
- **domain:** use validators in path constructors
- **domain:** move MustParsePath to testing.go
- **domain:** improve TraversalFreeValidator implementation
- **domain:** move all domain types to internal/domain
- **domain:** update all internal package imports to use internal/domain
- **domain:** use TargetPath for operation targets
- **domain:** format code and fix linter issues
- **hooks:** eliminate duplicate test run in pre-commit
- **path:** remove Path generic wrapper to eliminate code quality issue
- **pkg:** simplify scanForOrphanedLinks method
- **pkg:** convert Client to facade pattern
- **pkg:** extract helper methods in DoctorService
- **pkg:** simplify DoctorWithScan method
- **test:** improve benchmark tests with proper error handling

### Style
- fix goimports formatting

### Test
- **client:** add exhaustive tests for increased coverage margin
- **client:** add edge case tests for coverage buffer
- **client:** add comprehensive tests for pkg/dot Client struct
- **config:** add marshal strategy interface tests
- **config:** add default value constant tests
- **config:** add configuration key constant tests
- **domain:** add path validator tests
- **domain:** add Result unwrap helper tests
- **domain:** add error helper tests
- **domain:** add permission constant tests
- **integration:** implement integration test categories
- **integration:** implement comprehensive integration test infrastructure

### Pull Requests
- Merge pull request [#25](https://github.com/yaklabco/dot/issues/25) from jamesainslie/feature-documentation-shiznickle
- Merge pull request [#24](https://github.com/yaklabco/dot/issues/24) from jamesainslie/docs-update-document-index
- Merge pull request [#23](https://github.com/yaklabco/dot/issues/23) from jamesainslie/docs-add-root-links
- Merge pull request [#22](https://github.com/yaklabco/dot/issues/22) from jamesainslie/integration-testing
- Merge pull request [#21](https://github.com/yaklabco/dot/issues/21) from jamesainslie/feature-tech-debt
- Merge pull request [#20](https://github.com/yaklabco/dot/issues/20) from jamesainslie/feature-implement-git-changelog
- Merge pull request [#19](https://github.com/yaklabco/dot/issues/19) from jamesainslie/feature-domain-refactor

### BREAKING CHANGE

```
(internal only): internal/api package removed.
This only affects code that directly imported internal/api, which
should not exist since it was an internal package.
```



<a name="v0.1.0"></a>
## v0.1.0 - 2025-10-07
### Build
- **make:** add buildvcs flag for reproducible builds
- **makefile:** add build infrastructure with semantic versioning
- **release:** add Homebrew tap integration

### Chore
- update README and clean up whitespace in planner files
- **ci:** ignore reviews directory
- **ci:** ignore reviews directory
- **ci:** remove args from golangci
- **ci:** move initial design docs to docs folder, and keep ignoring them
- **ci:** ignore control files
- **deps:** update go.mod dependency classification
- **docs:** add planning docs
- **init:** initialize Go module and project structure

### Ci
- **github:** add GitHub Actions workflows and goreleaser configuration
- **lint:** replace golangci-lint-action with direct installation for v2.x compatibility
- **lint:** update golangci-lint version to v2.5.0
- **release:** install golangci-lint before running linters
- **release:** use GORELEASER_TOKEN for tap updates

### Docs
- mark core implementation complete
- replace tabs with spaces in Makefile code block
- complete with ignore patterns and scanner
- complete planner foundation with 100% coverage
- implement comprehensive documentation suite
- add and completion documents
- add completion summary and update changelog
- mark complete
- add completion documentation
- add completion summary
- add implementation plan for API enhancements
- add completion summary and update changelog
- add completion summary and update changelog
- add completion summary and update changelog
- add final implementation summary
- document code review improvements
- **adr:** add ADR-003 and ADR-004 for future enhancements
- **api:** remove GNU symlink reference from package documentation
- **api:** document Doctor breaking change and migration path
- **changelog:** update changelog for completion
- **changelog:** update with features and fixes
- **cli:** remove GNU symlink references from user-facing text
- **config:** add configuration guide and update README
- **config:** add configuration management design
- **dot:** enhance Result monad documentation with usage guidance
- **dot:** clarify ScanConfig field behavior in comments
- **executor:** update completion document with improved coverage
- **executor:** add completion document
- **install:** add Homebrew installation guide and release process
- **plan:** update plan with new CLI verb terminology
- **planner:** add implementation plan
- **planner:** document completion
- **plans:** add language hints to code blocks and fix formatting
- **readme:** update documentation for completion
- **review:** add code review remediation progress tracking
- **review:** add final coverage status and analysis
- **review:** add final remediation summary
- **review:** add language identifier to commit list code block
- **terminology:** adopt manage/unmanage/remanage command naming

### Feat
- **adapters:** implement slog logger and no-op adapters
- **adapters:** implement OS filesystem adapter
- **api:** implement Unmanage, Remanage, and Adopt operations
- **api:** add foundational types for Client API
- **api:** define Client interface for public API
- **api:** implement Client with Manage operation
- **api:** add comprehensive tests and documentation
- **api:** implement directory extraction and link set optimization
- **api:** update Doctor API to accept ScanConfig parameter
- **api:** update Doctor API to accept ScanConfig parameter
- **api:** implement incremental remanage with hash-based change detection
- **api:** add depth calculation and directory skip logic
- **api:** implement link count extraction from plan
- **api:** add DoctorWithScan for explicit scan configuration
- **api:** wire up orphaned link detection with safety limits
- **cli:** implement list command for package inventory
- **cli:** add scan control flags to doctor command
- **cli:** implement help system with examples and completion
- **cli:** implement progress indicators for operation feedback
- **cli:** implement terminal styling and layout system
- **cli:** implement output renderer infrastructure
- **cli:** add minimal CLI entry point for build validation
- **cli:** add config command for XDG configuration management
- **cli:** implement error formatting foundation for
- **cli:** implement status command for installation state inspection
- **cli:** implement doctor command for health checks
- **cli:** implement CLI infrastructure with core commands
- **cli:** implement UX polish with output formatting
- **cli:** implement command handlers for manage, unmanage, remanage, adopt
- **cli:** show complete operation breakdown in table summary
- **config:** wire backup directory through system
- **config:** implement extended configuration infrastructure
- **config:** implement configuration management with Viper and XDG compliance
- **config:** add Config struct with validation
- **domain:** implement operation type hierarchy
- **domain:** implement error taxonomy with user-facing messages
- **domain:** implement Result monad for functional error handling
- **domain:** implement phantom-typed paths for compile-time safety
- **domain:** implement domain value objects
- **domain:** add package-operation mapping to Plan
- **dot:** add ScanConfig types for orphaned link detection
- **executor:** add metrics instrumentation wrapper
- **executor:** implement parallel batch execution
- **executor:** implement executor with two-phase commit
- **ignore:** implement pattern matching engine and ignore sets
- **manifest:** implement FSManifestStore persistence
- **manifest:** add core manifest domain types
- **manifest:** implement content hashing for packages
- **manifest:** define ManifestStore interface
- **manifest:** implement manifest validation
- **operation:** add Execute and Rollback methods to operations
- **pipeline:** track package ownership in operation plans
- **pipeline:** surface conflicts and warnings in plan metadata
- **pipeline:** enhance context cancellation handling in pipeline stages
- **pipeline:** implement symlink pipeline with scanning, planning, resolution, and sorting stages
- **planner:** implement suggestion generation and conflict enrichment
- **planner:** implement conflict detection for links and directories
- **planner:** define conflict type enumeration
- **planner:** implement real desired state computation
- **planner:** implement desired state computation foundation
- **planner:** define resolution status types
- **planner:** implement resolve result type
- **planner:** implement conflict value object
- **planner:** implement resolution policy types and basic policies
- **planner:** implement main resolver function and policy dispatcher
- **planner:** integrate resolver with planning pipeline
- **planner:** implement dependency graph construction
- **planner:** implement parallelization analysis
- **planner:** implement topological sort with cycle detection
- **ports:** define infrastructure port interfaces
- **scanner:** implement tree scanning with recursive traversal
- **scanner:** implement dotfile translation logic
- **scanner:** implement package scanner with ignore support
- **types:** add Status and PackageInfo types

### Fix
- **api:** address CodeRabbit feedback on
- **api:** use configured skip patterns in recursive orphan scanning
- **api:** improve error handling and test robustness
- **api:** use package-operation mapping for accurate manifest tracking
- **api:** normalize paths for cross-platform link lookup
- **api:** enforce depth and context limits in recursive orphan scanning
- **cli:** resolve critical bugs in progress, config, and rendering
- **cli:** handle both pointer and value operation types in renderers
- **cli:** improve config format detection and help text indentation
- **cli:** correct scan flag variable scope in NewDoctorCommand
- **cli:** add error templates for checkpoint and not implemented errors
- **cli:** respect NO_COLOR environment variable in shouldColorize
- **cli:** improve JSON/YAML output and doctor performance
- **cli:** render execution plan in dry-run mode
- **cli:** improve TTY detection portability and path truncation
- **config:** enable CodeRabbit auto-review for all pull requests
- **executor:** make Checkpoint operations map thread-safe
- **executor:** address code review feedback for concurrent safety and error handling
- **manifest:** add security guards and prevent hash collisions
- **pipeline:** prevent shared mutation of context maps in metadata conversion
- **release:** separate archive configs for Homebrew compatibility
- **scanner:** implement real package tree scanning with ignore filtering
- **test:** improve test isolation and cross-platform compatibility
- **test:** make Adopt execution error test deterministic

### Refactor
- **adopt:** update Adopt and PlanAdopt methods to use files-first signature
- **api:** reduce cyclomatic complexity in PlanRemanage
- **api:** extract orphan scan logic to reduce complexity
- **cli:** address code review nitpicks for improved code quality
- **cli:** reduce cyclomatic complexity in table renderer
- **cli:** add default case and eliminate type assertion duplication
- **pipeline:** use safe unwrap pattern in path construction tests
- **pipeline:** improve test quality and organization
- **quality:** improve error handling documentation and panic messages
- **terminology:** update suggestion text from unmanage to unmanage
- **terminology:** replace symlink with package directory terminology
- **terminology:** complete symlink removal from test fixtures
- **terminology:** rename symlink-prefixed variables to package/manage

### Style
- **all:** apply goimports formatting
- **all:** apply goimports formatting
- **domain:** fix linting issues and apply formatting
- **manifest:** apply goimports formatting
- **planner:** fix linting issues in implementation
- **scanner:** apply goimports formatting
- **test:** format test files with goimports

### Test
- **adapters:** add comprehensive MemFS tests to achieve 80%+ coverage
- **api:** add manifest helper tests and document remediation
- **api:** add comprehensive test coverage for all API methods
- **cli:** fix help text assertion after symlink removal
- **cli:** increase cmd/dot test coverage to 88.6%
- **cli:** add comprehensive tests to restore coverage above 80%
- **cmd:** add basic command constructor tests
- **config:** add comprehensive loader and precedence tests
- **config:** add aggressive coverage boost tests
- **config:** add validation edge case tests
- **config:** improve test coverage to 83%
- **coverage:** increase test coverage from 73.8% to 83.7%
- **dot:** add comprehensive error and operation tests
- **executor:** add comprehensive tests to exceed 80% coverage threshold
- **planner:** add coverage tests to exceed 80 percent threshold

### Pull Requests
- Merge pull request [#18](https://github.com/yaklabco/dot/issues/18) from jamesainslie/feature-homebrew-tap
- Merge pull request [#17](https://github.com/yaklabco/dot/issues/17) from jamesainslie/fix-dry-run-output
- Merge pull request [#16](https://github.com/yaklabco/dot/issues/16) from jamesainslie/feature-implement-stubs
- Merge pull request [#15](https://github.com/yaklabco/dot/issues/15) from jamesainslie/feature-remove-symlink-terminology
- Merge pull request [#14](https://github.com/yaklabco/dot/issues/14) from jamesainslie/feature-remove-symlink-references
- Merge pull request [#13](https://github.com/yaklabco/dot/issues/13) from jamesainslie/feature-api-enhancements
- Merge pull request [#12](https://github.com/yaklabco/dot/issues/12) from jamesainslie/feature-code-review-remediation
- Merge pull request [#11](https://github.com/yaklabco/dot/issues/11) from jamesainslie/feature-error-handling-ux
- Merge pull request [#10](https://github.com/yaklabco/dot/issues/10) from jamesainslie/feature-implement-cli-query
- Merge pull request [#9](https://github.com/yaklabco/dot/issues/9) from jamesainslie/feature-implement-cli
- Merge pull request [#7](https://github.com/yaklabco/dot/issues/7) from jamesainslie/feature-implement-api
- Merge pull request [#6](https://github.com/yaklabco/dot/issues/6) from jamesainslie/feature-manifests-state-management
- Merge pull request [#5](https://github.com/yaklabco/dot/issues/5) from jamesainslie/feature-executor
- Merge pull request [#4](https://github.com/yaklabco/dot/issues/4) from jamesainslie/jamesainslie-implement-pipeline-orchestration
- Merge pull request [#3](https://github.com/yaklabco/dot/issues/3) from jamesainslie/jamesainslie-implement-topological-sorter
- Merge pull request [#2](https://github.com/yaklabco/dot/issues/2) from jamesainslie/jamesainslie-implement-resolver
- Merge pull request [#1](https://github.com/yaklabco/dot/issues/1) from jamesainslie/jamesainslie-implement-func-scanner


[Unreleased]: https://github.com/yaklabco/dot/compare/v0.5.0...HEAD
[v0.5.0]: https://github.com/yaklabco/dot/compare/v0.4.4...v0.5.0
[v0.4.4]: https://github.com/yaklabco/dot/compare/v0.4.3...v0.4.4
[v0.4.3]: https://github.com/yaklabco/dot/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/yaklabco/dot/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/yaklabco/dot/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/yaklabco/dot/compare/v0.3.1...v0.4.0
[v0.3.1]: https://github.com/yaklabco/dot/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/yaklabco/dot/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/yaklabco/dot/compare/v0.1.1...v0.2.0
[v0.1.1]: https://github.com/yaklabco/dot/compare/v0.1.0...v0.1.1
