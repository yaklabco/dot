# Release Workflow Diagram

## Overview

```mermaid
flowchart TD
    A[Developer Action] --> B{Choose Method}
    B -->|Option 1| C[GitHub UI: Actions → Version Bump → Run]
    B -->|Option 2| D[CLI: gh workflow run version-bump.yml]
    B -->|Option 3| E[Local: make version-patch && git push]
    
    C --> F[Version Bump Workflow]
    D --> F
    E --> F
    
    F --> G[1. Checkout with full history]
    G --> H[2. Setup: Go + git-chglog + golangci-lint]
    H --> I[3. Calculate next version]
    I --> J[4. Run quality checks]
    J --> K{All checks pass?}
    K -->|No| L[ Workflow fails - fix issues]
    K -->|Yes| M[5. Generate CHANGELOG.md]
    M --> N[6. Commit CHANGELOG.md to main]
    N --> O[7. Create tag]
    O --> P[8. Regenerate changelog with tag]
    P --> Q[9. Amend commit]
    Q --> R[10. Move tag to amended commit]
    R --> S[11. Push main branch]
    S --> T[12. Push tag]
    
    T -->|Tag push triggers| U[Release Workflow]
    
    U --> V[1. Checkout with full history]
    V --> W[2. Setup: Go + golangci-lint]
    W --> X[3. Run tests]
    X --> Y[4. Run linters]
    Y --> Z{Tests & linters pass?}
    Z -->|No| AA[ Release fails]
    Z -->|Yes| AB[5. Extract changelog section]
    AB --> AC[6. Run GoReleaser]
    AC --> AD[Build binaries for all platforms]
    AD --> AE[Create archives]
    AE --> AF[Generate checksums]
    AF --> AG[Create GitHub Release]
    AG --> AH[Upload artifacts]
    AH --> AI[Update Homebrew tap]
    AI --> AJ[ Release Complete!]
    
    style F fill:#e1f5ff
    style U fill:#e1f5ff
    style AJ fill:#90EE90
    style L fill:#ffcccc
    style AA fill:#ffcccc
```

## Sequential Flow

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant GH as GitHub UI/CLI
    participant VB as Version Bump Workflow
    participant Git as Git Repository
    participant RW as Release Workflow
    participant GR as GoReleaser
    participant Brew as Homebrew Tap
    
    Dev->>GH: Trigger version bump (patch/minor/major)
    GH->>VB: Start workflow
    
    VB->>VB: Run quality checks
    Note over VB: Tests, Coverage, Linting
    
    VB->>VB: Calculate next version
    VB->>VB: Generate CHANGELOG.md
    VB->>Git: Commit changelog
    VB->>Git: Create & push tag (v0.4.4)
    Git->>VB:  Tag pushed
    
    Note over Git,RW: Tag push triggers Release Workflow
    
    Git->>RW: Trigger on tag: v0.4.4
    RW->>RW: Run tests & linters
    RW->>RW: Extract changelog section
    RW->>GR: Execute GoReleaser
    
    par Build all platforms
        GR->>GR: Build Linux amd64
        GR->>GR: Build Linux arm64
        GR->>GR: Build macOS amd64
        GR->>GR: Build macOS arm64
        GR->>GR: Build Windows amd64
    end
    
    GR->>GR: Create archives & checksums
    GR->>Git: Create GitHub Release
    GR->>Brew: Update formula
    
    RW->>Dev:  Release complete!
    Note over Dev: v0.4.4 available for download
```

## Parallel Build Process

```mermaid
graph TD
    A[GoReleaser Starts] --> B{Build Manager}
    
    B -->|Parallel| C[Linux amd64]
    B -->|Parallel| D[Linux arm64]
    B -->|Parallel| E[macOS amd64]
    B -->|Parallel| F[macOS arm64]
    B -->|Parallel| G[Windows amd64]
    
    C --> H[Collect Binaries]
    D --> H
    E --> H
    F --> H
    G --> H
    
    H --> I[Create Archives]
    I --> J[.tar.gz for Unix]
    I --> K[.zip for Windows]
    
    J --> L[Generate Checksums]
    K --> L
    
    L --> M[Upload to GitHub Release]
    M --> N[Update Homebrew Tap]
    N --> O[ Complete]
    
    style A fill:#e1f5ff
    style O fill:#90EE90
```

## Quality Gates

```mermaid
flowchart LR
    subgraph VB [Version Bump Workflow Gates]
        A1[Run Tests] --> A2{Pass?}
        A2 -->|No| FAIL1[ Fail]
        A2 -->|Yes| A3[Check Coverage ≥80%]
        A3 --> A4{Pass?}
        A4 -->|No| FAIL1
        A4 -->|Yes| A5[Run golangci-lint]
        A5 --> A6{Pass?}
        A6 -->|No| FAIL1
        A6 -->|Yes| A7[Run go vet]
        A7 --> A8{Pass?}
        A8 -->|No| FAIL1
        A8 -->|Yes| PASS1[ Create Tag]
    end
    
    PASS1 --> TRIGGER[Tag Push]
    
    subgraph RW [Release Workflow Gates]
        B1[Run Tests] --> B2{Pass?}
        B2 -->|No| FAIL2[ Fail]
        B2 -->|Yes| B3[Run golangci-lint]
        B3 --> B4{Pass?}
        B4 -->|No| FAIL2
        B4 -->|Yes| PASS2[ Build & Release]
    end
    
    TRIGGER --> RW
    
    style PASS1 fill:#90EE90
    style PASS2 fill:#90EE90
    style FAIL1 fill:#ffcccc
    style FAIL2 fill:#ffcccc
```

## Configuration Files Structure

```mermaid
graph TD
    A[dot Repository] --> B[.github/]
    A --> C[.chglog/]
    A --> D[Configuration Files]
    A --> E[docs/]
    
    B --> B1[workflows/]
    B --> B2[RELEASE.md 🆕]
    B --> B3[WORKFLOW_DIAGRAM.md 🆕]
    
    B1 --> B1A[version-bump.yml 🆕]
    B1 --> B1B[release.yml Updated]
    B1 --> B1C[ci.yml]
    
    C --> C1[config.yml]
    C --> C2[CHANGELOG.tpl.md]
    
    D --> D1[.goreleaser.yml Updated]
    D --> D2[Makefile]
    D --> D3[CHANGELOG.md Auto-generated]
    
    E --> E1[developer/]
    E1 --> E1A[release-workflow.md Updated]
    
    style B1A fill:#90EE90
    style B2 fill:#90EE90
    style B3 fill:#90EE90
    style B1B fill:#FFD700
    style D1 fill:#FFD700
    style E1A fill:#FFD700
    style D3 fill:#87CEEB
```

## Before vs After Comparison

```mermaid
flowchart TD
    subgraph BEFORE [Before: Manual Process]
        direction TB
        M1[Developer runs make version-patch]
        M1 --> M2[Run tests locally]
        M2 --> M3[Run linting locally]
        M3 --> M4[Generate changelog locally]
        M4 --> M5[Create tag locally]
        M5 --> M6[Manual: git push origin main]
        M6 --> M7[Manual: git push origin tag]
        M7 --> M8[GitHub Actions: Release Workflow]
        M8 --> M9[Build binaries]
        M9 --> M10[Create release]
        M10 --> M11[Update Homebrew]
        
        M12[Time: 5-10 minutes<br/>Manual steps: 3-4<br/>Error prone: Medium]
    end
    
    subgraph AFTER [After: Automated Process]
        direction TB
        A1[Developer clicks Run Workflow]
        A1 --> A2[GitHub Actions: Version Bump]
        A2 --> A3[Quality checks]
        A3 --> A4[Generate changelog]
        A4 --> A5[Create tag]
        A5 --> A6[Auto-push to GitHub]
        A6 --> A7[GitHub Actions: Release Workflow]
        A7 --> A8[Build binaries]
        A8 --> A9[Create release]
        A9 --> A10[Update Homebrew]
        
        A11[Time: 3-5 minutes<br/>Manual steps: 1<br/>Error prone: Low]
    end
    
    style BEFORE fill:#ffe6e6
    style AFTER fill:#e6ffe6
    style M1 fill:#ffcccc
    style A1 fill:#90EE90
```

## Rollback Process

```mermaid
flowchart TD
    A[Need to Rollback Release v0.4.4] --> B[1. Delete GitHub Release]
    B --> C[gh release delete v0.4.4 --yes]
    C --> D[2. Delete Remote Tag]
    D --> E[git push origin :refs/tags/v0.4.4]
    E --> F[3. Delete Local Tag]
    F --> G[git tag -d v0.4.4]
    G --> H{Revert changelog commit?}
    H -->|Yes| I[git revert HEAD]
    I --> J[git push origin main]
    H -->|No| K[Keep changelog as-is]
    J --> L[4. Fix Issues]
    K --> L
    L --> M[Fix bugs/issues that caused rollback]
    M --> N[5. Re-release]
    N --> O[Run version bump workflow again]
    O --> P[ New release created]
    
    style A fill:#ffcccc
    style P fill:#90EE90
```

## Troubleshooting Flowchart

```mermaid
flowchart TD
    A[Workflow Failed?] --> B{What type of failure?}
    
    B -->|Quality Checks| C[Quality Checks Failed]
    C --> C1{Which check?}
    C1 -->|Tests| C2[Fix failing tests]
    C1 -->|Coverage| C3[Add tests to reach 80%]
    C1 -->|Linting| C4[Fix linting errors]
    C1 -->|go vet| C5[Fix vet issues]
    C2 --> C6[Commit & push fixes]
    C3 --> C6
    C4 --> C6
    C5 --> C6
    C6 --> C7[Re-run workflow]
    
    B -->|Tag Exists| D[Tag Already Exists]
    D --> D1[Delete local tag: git tag -d vX.Y.Z]
    D1 --> D2[Delete remote tag: git push origin :refs/tags/vX.Y.Z]
    D2 --> D3[Re-run workflow]
    
    B -->|Permissions| E[Permission Error]
    E --> E1[Check Settings → Actions → General]
    E1 --> E2[Enable Read and write permissions]
    E2 --> E3[Re-run workflow]
    
    B -->|GoReleaser| F[GoReleaser Failed]
    F --> F1{Which component?}
    F1 -->|Homebrew| F2[Check GITHUB_TOKEN secret]
    F1 -->|Build| F3[Check build configuration]
    F1 -->|Assets| F4[Check file paths]
    F2 --> F5[Fix configuration]
    F3 --> F5
    F4 --> F5
    F5 --> F6[Delete tag and re-run]
    
    C7 --> G[ Success]
    D3 --> G
    E3 --> G
    F6 --> G
    
    style A fill:#ffcccc
    style G fill:#90EE90
```

## Version Bump Decision Tree

```mermaid
flowchart TD
    A[Ready to Release?] --> B{What kind of changes?}
    
    B -->|Bug fixes<br/>Small improvements<br/>No new features| C[PATCH Release]
    C --> C1[v0.4.3 → v0.4.4]
    C1 --> C2[Select: patch]
    
    B -->|New features<br/>Backward compatible<br/>No breaking changes| D[MINOR Release]
    D --> D1[v0.4.3 → v0.5.0]
    D1 --> D2[Select: minor]
    
    B -->|Breaking changes<br/>API changes<br/>Major refactoring| E[MAJOR Release]
    E --> E1[v0.4.3 → v1.0.0]
    E1 --> E2[Select: major]
    
    C2 --> F[Run Version Bump Workflow]
    D2 --> F
    E2 --> F
    
    F --> G{Conventional commits?}
    G -->|Yes| H[Changelog generated correctly]
    G -->|No| I[ Changelog may be incomplete]
    I --> J[Update commit messages]
    J --> F
    
    H --> K[ Release Published]
    
    style C fill:#e1f5ff
    style D fill:#fff4e1
    style E fill:#ffe1e1
    style K fill:#90EE90
```

## Monitoring Workflow

```mermaid
flowchart LR
    A[Monitor Release] --> B{Choose Tool}
    
    B -->|GitHub UI| C[Visit Actions Tab]
    C --> C1[Select workflow run]
    C1 --> C2[View logs]
    
    B -->|GitHub CLI| D[Use gh commands]
    D --> D1[gh run list]
    D --> D2[gh run watch]
    D --> D3[gh run view --log]
    
    B -->|Releases| E[Check releases]
    E --> E1[gh release list]
    E --> E2[gh release view vX.Y.Z]
    
    C2 --> F[Monitor Progress]
    D3 --> F
    E2 --> G[Verify Release]
    
    F --> H{Success?}
    H -->|Yes| I[ Done]
    H -->|No| J[Check logs & fix]
    J --> K[Re-run workflow]
    
    G --> L{Artifacts OK?}
    L -->|Yes| I
    L -->|No| M[Report issue]
    
    style I fill:#90EE90
    style J fill:#ffcccc
    style M fill:#ffcccc
```

## Legend

```mermaid
flowchart LR
    A[Regular Step] --> B[Next Step]
    C{Decision Point} -->|Option| D[Action]
    
    subgraph Status
        E[🆕 New File]
        F[ Updated File]
        G[ Success]
        H[ Failure]
        I[ Warning]
    end
    
    style E fill:#90EE90
    style F fill:#FFD700
    style G fill:#90EE90
    style H fill:#ffcccc
    style I fill:#fff4cc
```

## Quick Reference Commands

| Action | GitHub UI | GitHub CLI |
|--------|-----------|------------|
| **Run patch release** | Actions → Version Bump → patch | `gh workflow run version-bump.yml -f version_type=patch` |
| **Run minor release** | Actions → Version Bump → minor | `gh workflow run version-bump.yml -f version_type=minor` |
| **Run major release** | Actions → Version Bump → major | `gh workflow run version-bump.yml -f version_type=major` |
| **Watch workflow** | Actions → View run | `gh run watch` |
| **View logs** | Actions → Run → Logs | `gh run view --log` |
| **List releases** | Releases tab | `gh release list` |
| **View release** | Click release | `gh release view vX.Y.Z` |
| **Download artifacts** | Release → Assets | `gh release download vX.Y.Z` |

## Best Practices Summary

```mermaid
mindmap
  root((Release<br/>Best Practices))
    Commits
      Use Conventional Commits
      Include scope when possible
      Clear descriptions
      BREAKING CHANGE for majors
    Testing
      Run CI workflow not local
      Test on branch first
      Monitor workflow logs
      Verify installation after
    Quality
      Maintain 80%+ coverage
      Fix linting before release
      No go vet warnings
      All tests pass
    Process
      Use automated workflow
      One click releases
      Review changelog
      Update docs if needed
    Monitoring
      Watch workflow progress
      Check GitHub Release
      Verify Homebrew tap
      Test binary downloads
```
