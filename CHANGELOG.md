
## [Unreleased]

## [0.1.0] - 2026-01-27

### Added
* 374bca3 improved devex for building locally
* 9aee443 fixing tarball name
* 34b49e7 fix build flags
* 6d313f5 set plugin value
* 45bf541 ci release uses taskfile now
* 6af200b bump version to test release process
* 8be263f wip reworking artifact installation and release process
* d2d2c92 fixed install script version check
* 8a08887 added version command
* a358139 fixed plugin version number
* 9bd0a52 set examples from footer comments
* 57e605b updated old project name
* 9f544a3 updated task tests
* 0bb3999 fixed a broken test
* 376be83 mostly fixed marshalling schema properties
* da5dcaf better organization or comment resources
* df7a796 improved template filename and new file for markup types
* a978d96 better function name for cmd configs
* 4a63f2e removed old file
* 1eb0a7d finished cleaning up cmd
* 2a36fa2 small todo list cleanup
* a598d23 massive codebase re-org
* 9bbed70 support values order in docs
* dede182 improved multi-line support in mardown table
* e4ee407 fix markdown rows for values with multi-line descriptions
* c2c5f99 fixed comment schema errors not being logged
* 838e909 warn about untyped values
* 6e106c7 fixed values tabe not being in a consistent order
* 378bb73 correctly parse comments where the header comment includes comments from the previous node
* 08f6a63 removed development output
* bf2f48c fixed null yaml values causing schema generation to fail
* 7160311 added note about supporting redhat.vscode-yaml tooltips
* f94a1e9 fixed something, i dont remember

## [0.0.2] - 2025-11-25

### Added
* 29075c2 removed unused schema-file cli param
* 033b814 reorganized dev roadmap
* 5ec65b9 small readme fixes
* ea00f45 support writing schema modeline to values.yaml
* bc6a585 trying to fix artifacthub entry
* a31148b updated cli help output in readme
* 8a2daa9 updated roadmap
* 71cc14b warn about undocumented fields
* 6846bed wip: documenting the schema comments

## [0.0.1] - 2025-11-24

### Added
* 3f70951 added artifact repo config
* 5699fff added layered fs to load os and embedded templates
* e091ee6 adding ci
* fa92a63 and also rename the ci workflow
* da1cfa5 attempted to improve logging
* 82ceb14 chart discovery is disconnected from Plan
* 993b9dd ci badge
* c787400 cleaned up comment parsing & added tests
* 2b1babd cleanup task helper
* 86f71db correctly name the artifacthub repo file
* 0214437 docs command
* 62dd0a3 fixed bin path on helm plugin
* 41d4530 fixed several issues with comment yaml parsing
* bf5b593 fixed tag trigger on ci release
* 4e9430e forgot to update roadmap in 86f40cb
* 6ff6658 goreleaser creates release?
* 86f40cb gracefully fail when no template available & use-default=false
* 82e4ae7 helm plugin support
* 0c4ab80 huge docs cli improvements
* 9674c1a implemented recursive chart discovery
* f096998 improved comment parsing error descriptions
* d7267c9 improved readme
* df1e1e1 initial commit
* 3381e78 load static templates before user specified templates
* 2da4c63 more goreleaser stuff, starting to test ci
* 0134241 more readme organization
* ec37b67 notes on dev roadmap
* eb3baa8 oh look, an artifacthub badge
* a7a1651 promote plugin to 0.0.1
* fd47ce5 re-enable jsonschema not,if,then,else properties
* d23a71b removed unecessary fs for file ops
* 51c237c renamed ci workflow
* 3c8350e renamed cmd
* 0428394 renamed release ci workflow
* 0ae056a revert snapshot use
* 690469a slight improvements to debug log formatting
* afda5e3 small readme tweak + removed commented code
* f3d5694 some roadmap items i forgot to check off
* fb04b7e take chart paths as command args instead of flags
* 1aae996 try to fix artifacthub warning
* ae49654 update readme name
* f3cda30 updated roadmap
* a9667f4 use snapshot builds until CI is sorted out
* 8ee2d87 wip formalizing cobra commands, made space for docs command
* 84c540e wip improving templating facilities
* 99079e3 wip: switching to goreleaser for building & packaging
* 238ee7b working through release issues...
