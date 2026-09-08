"""Language registry for per-SDK versioning and release tags."""

from __future__ import annotations

import json
import re
import subprocess
from dataclasses import dataclass
from pathlib import Path

from commits import HEADER_RE, MERGE_RE, REVERT_RE, Commit, bump_level, log, resolve

REGISTRY_FILE = Path(__file__).with_name("languages.json")
CHANGELOG_FILE = "CHANGELOG.md"
SEMVER = re.compile(r"^\d+\.\d+\.\d+$")


class RegistryError(Exception):
    pass


@dataclass(frozen=True)
class Language:
    id: str
    name: str
    dir: str
    tag_prefix: str
    marker: str
    scopes: tuple[str, ...]

    def version_path(self, root: Path) -> Path:
        return root / self.dir / "VERSION"

    def tag_name(self, version: str) -> str:
        return f"{self.tag_prefix}/v{version}"

    def changelog_path(self, root: Path) -> Path:
        return root / self.dir / CHANGELOG_FILE

    def changelog_heading(self, version: str) -> str:
        return f"## [{version}]"

    def matches_scope(self, scope: str | None) -> bool:
        if not scope:
            return False
        return scope == self.id or scope in self.scopes

    def matches_path(self, path: str) -> bool:
        prefix = f"{self.dir}/"
        return path == self.dir or path.startswith(prefix)

    def is_present(self, root: Path) -> bool:
        return (root / self.marker).is_file()

    def is_releasable(self, root: Path) -> bool:
        return self.version_path(root).is_file()


def load_registry(path: Path | None = None) -> list[Language]:
    data = json.loads((path or REGISTRY_FILE).read_text(encoding="utf-8"))
    return [
        Language(
            id=item["id"],
            name=item["name"],
            dir=item["dir"],
            tag_prefix=item["tag_prefix"],
            marker=item["marker"],
            scopes=tuple(item["scopes"]),
        )
        for item in data["languages"]
    ]


def language_by_id(lang_id: str, registry: list[Language] | None = None) -> Language:
    for lang in registry or load_registry():
        if lang.id == lang_id:
            return lang
    raise RegistryError(f"unknown language: {lang_id!r}")


def _git(*args: str, cwd: Path) -> str:
    return subprocess.run(
        ("git", *args), cwd=cwd, check=True, text=True, capture_output=True
    ).stdout


def semver_tuple(version: str) -> tuple[int, int, int]:
    return tuple(map(int, version.split(".")))


def read_version(lang: Language, root: Path) -> str:
    path = lang.version_path(root)
    if not path.is_file():
        raise RegistryError(f"missing {path.relative_to(root)}")
    version = path.read_text(encoding="utf-8").strip()
    if not SEMVER.match(version):
        raise RegistryError(f"invalid semver in {path}: {version!r}")
    return version


def version_at_ref(lang: Language, ref: str, *, root: Path) -> str | None:
    rel = lang.version_path(root).relative_to(root).as_posix()
    try:
        text = _git("show", f"{ref}:{rel}", cwd=root).strip()
    except subprocess.CalledProcessError:
        return None
    return text if SEMVER.match(text) else None


def changelog_has(lang: Language, version: str, *, root: Path) -> bool:
    path = lang.changelog_path(root)
    if not path.is_file():
        return False
    heading = lang.changelog_heading(version)
    return heading in path.read_text(encoding="utf-8")


def _commit_files(sha: str, *, root: Path) -> list[str]:
    out = _git("show", "--name-only", "--format=", sha, cwd=root)
    return [line for line in out.splitlines() if line.strip()]


def _release_worthy(commit: Commit) -> bool:
    if MERGE_RE.match(commit.subject) or REVERT_RE.match(commit.subject):
        return False
    match = HEADER_RE.match(commit.subject)
    if not match:
        return False
    if match.group("breaking") or "BREAKING CHANGE" in commit.body:
        return True
    return match.group("type") in {"feat", "fix", "perf"}


def languages_for_commit(
    commit: Commit, registry: list[Language], *, root: Path
) -> set[str]:
    langs: set[str] = set()
    match = HEADER_RE.match(commit.subject)
    if match and match.group("scope"):
        scope = match.group("scope")
        for lang in registry:
            if lang.matches_scope(scope):
                langs.add(lang.id)
    for path in _commit_files(commit.sha, root=root):
        for lang in registry:
            if lang.matches_path(path):
                langs.add(lang.id)
    return langs


def commits_in_range(base: str, head: str, *, root: Path) -> list[Commit]:
    resolve(base, cwd=root)
    resolve(head, cwd=root)
    return log(f"{base}..{head}", cwd=root)


def release_plan(
    base: str, head: str, *, root: Path | None = None
) -> dict[str, list[Commit]]:
    root = root or Path.cwd()
    registry = load_registry()
    plan: dict[str, list[Commit]] = {lang.id: [] for lang in registry}
    for commit in commits_in_range(base, head, root=root):
        if not _release_worthy(commit):
            continue
        for lang_id in languages_for_commit(commit, registry, root=root):
            plan[lang_id].append(commit)
    return {lang_id: commits for lang_id, commits in plan.items() if commits}


def check_release_mr(base: str, head: str, *, root: Path | None = None) -> str | None:
    root = root or Path.cwd()
    registry = load_registry()
    plan = release_plan(base, head, root=root)

    for lang in registry:
        commits = plan.get(lang.id, [])
        needed = bump_level(commits) if commits else None
        base_version = version_at_ref(lang, base, root=root)
        head_path = lang.version_path(root)
        head_version = (
            head_path.read_text(encoding="utf-8").strip()
            if head_path.is_file()
            else None
        )

        if base_version is None and head_version is None:
            if commits:
                return (
                    f"{lang.id}: release-worthy commits but no {head_path.relative_to(root)}; "
                    f"add VERSION when the {lang.name} SDK is ready to ship"
                )
            continue

        if head_version is None:
            if commits:
                return f"{lang.id}: missing {head_path.relative_to(root)}"
            continue

        if not SEMVER.match(head_version):
            return f"{lang.id}: invalid semver in {head_path.relative_to(root)}: {head_version!r}"

        if base_version is None:
            base_version = "0.0.0"

        if needed:
            if semver_tuple(head_version) <= semver_tuple(base_version):
                return (
                    f"{lang.id}: release-worthy commits (expected {needed} bump) but "
                    f"{head_path.relative_to(root)} is still {head_version} "
                    f"(was {base_version})"
                )
            if not changelog_has(lang, head_version, root=root):
                rel = lang.changelog_path(root).relative_to(root)
                return (
                    f"{lang.id}: {head_path.relative_to(root)} is {head_version} but "
                    f"{rel} has no {lang.changelog_heading(head_version)!r} section"
                )
            continue

        if semver_tuple(head_version) > semver_tuple(
            base_version
        ) and not changelog_has(lang, head_version, root=root):
            rel = lang.changelog_path(root).relative_to(root)
            return (
                f"{lang.id}: {head_path.relative_to(root)} bumped to {head_version} but "
                f"{rel} has no {lang.changelog_heading(head_version)!r} section"
            )

    return None
