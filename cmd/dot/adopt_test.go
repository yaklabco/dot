package main

// The following test functions were removed as part of the adopt command simplification:
//
// - TestCommonPrefix: tested commonPrefix() helper function
// - TestDeriveCommonPackageName: tested deriveCommonPackageName() helper function
// - TestFileExists: tested fileExists() helper function
//
// These functions were removed when glob mode auto-detection was eliminated from the adopt
// command. The new behavior requires explicit package names when adopting multiple files:
//
//   dot adopt .ssh              # Auto-naming for single file
//   dot adopt vim .vimrc .vim   # Explicit package for multiple files
//
// The adopt command now has only two modes: single-file auto-naming and explicit multi-file.
// This simplification removes ambiguous behavior and makes the command more predictable.
