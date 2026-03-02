#!/usr/bin/env bash
# verify-deps.sh — Belt-and-suspenders import violation checker (complements depguard)
# Usage: ./scripts/verify-deps.sh
set -euo pipefail

FAIL=0

echo "==> Checking Clean Architecture import violations..."

# DEP-2: Domain must not import usecase, adapter, or infrastructure
if grep -rn '"telegram-trello-bot/internal/usecase\|telegram-trello-bot/internal/adapter\|telegram-trello-bot/internal/infrastructure' internal/domain/ 2>/dev/null | grep -v "_test.go"; then
    echo "FAIL: domain/ has forbidden imports (violates DEP-1, DEP-2)"
    FAIL=1
fi

# DEP-1: Use cases must not import adapter or infrastructure
if grep -rn '"telegram-trello-bot/internal/adapter\|telegram-trello-bot/internal/infrastructure' internal/usecase/ 2>/dev/null | grep -v "_test.go"; then
    echo "FAIL: usecase/ has forbidden imports (violates DEP-1)"
    FAIL=1
fi

# DEP-1: Infrastructure must not import usecase (except usecase/port)
if grep -rn '"telegram-trello-bot/internal/usecase/' internal/infrastructure/ 2>/dev/null | grep -v "_test.go" | grep -v "usecase/port"; then
    echo "FAIL: infrastructure/ imports usecase (not port) (violates DEP-1)"
    FAIL=1
fi

# FORBID-1: No global mutable state (var declarations that aren't errors or regexps)
if grep -rn '^var ' internal/ --include="*.go" 2>/dev/null | grep -v "_test.go" | grep -v "Err" | grep -v "regexp" | grep -v "Regex\|regex"; then
    echo "WARN: Found global var declarations — verify they are not mutable state (FORBID-1)"
fi

# FORBID-2: No init() functions
if grep -rn 'func init()' internal/ --include="*.go" 2>/dev/null | grep -v "_test.go"; then
    echo "FAIL: Found init() functions (violates FORBID-2)"
    FAIL=1
fi

# ERR-3: No panics in business logic
if grep -rn 'panic(' internal/ --include="*.go" 2>/dev/null | grep -v "_test.go"; then
    echo "FAIL: Found panic() in business logic (violates ERR-3)"
    FAIL=1
fi

if [ $FAIL -eq 1 ]; then
    echo "==> Dependency verification FAILED"
    exit 1
fi

echo "==> All dependency checks passed."
