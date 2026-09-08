"""Shared conventional-commit parsing and git log helpers."""

from __future__ import annotations

import re
import subprocess
from collections.abc import Iterable
from dataclasses import dataclass
from pathlib import Path

HEADER_RE = re.compile(
    r"^(?P<type>build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)"
    r"(?:\((?P<scope>[^)]+)\))?(?P<breaking>!)?: (?P<subject>.+)$"
)
MERGE_RE = re.compile(r"^Merge (branch|commit|remote-tracking branch|pull request) ")
REVERT_RE = re.compile(r"^Revert ")
HINT = "expected: type(scope): subject  (e.g. feat(go): add provider)"


@dataclass(frozen=True)
class Commit:
    sha: str
    subject: str
    body: str


class GitError(Exception):
    pass


def git(*args: str, cwd: Path | None = None) -> str:
    try:
        return subprocess.run(
            ("git", *args), cwd=cwd, check=True, text=True, capture_output=True
        ).stdout
    except subprocess.CalledProcessError as exc:
        raise GitError(f"git {' '.join(args)} failed") from exc


def resolve(ref: str, *, cwd: Path | None = None) -> None:
    git("rev-parse", "--verify", "--quiet", f"{ref}^{{commit}}", cwd=cwd)


def log(range_spec: str, *, cwd: Path | None = None) -> list[Commit]:
    commits: list[Commit] = []
    for entry in git("log", "--format=%H%x1f%s%x1f%b%x1e", range_spec, cwd=cwd).split(
        "\x1e"
    ):
        entry = entry.strip("\n")
        if not entry:
            continue
        sha, subject, body = entry.split("\x1f", 2)
        commits.append(Commit(sha=sha, subject=subject, body=body))
    return commits


def validate_subject(subject: str) -> str | None:
    if MERGE_RE.match(subject) or REVERT_RE.match(subject):
        return None
    if not HEADER_RE.match(subject):
        return f"non-conventional subject: {subject!r}"
    return None


def validate_range(base: str, head: str, *, cwd: Path | None = None) -> str | None:
    resolve(base, cwd=cwd)
    resolve(head, cwd=cwd)
    for commit in log(f"{base}..{head}", cwd=cwd):
        if err := validate_subject(commit.subject):
            return f"{commit.sha[:8]}: {err}"
    return None


def bump_level(commits: Iterable[Commit]) -> str | None:
    level: str | None = None
    for commit in commits:
        if MERGE_RE.match(commit.subject) or REVERT_RE.match(commit.subject):
            continue
        match = HEADER_RE.match(commit.subject)
        if not match:
            continue
        if match.group("breaking") or "BREAKING CHANGE" in commit.body:
            return "major"
        typ = match.group("type")
        if typ == "feat":
            level = "minor"
        elif typ in {"fix", "perf"} and level is None:
            level = "patch"
    return level


def bump_between(base: str, head: str, *, cwd: Path | None = None) -> str | None:
    return bump_level(log(f"{base}..{head}", cwd=cwd))
