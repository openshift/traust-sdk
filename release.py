#!/usr/bin/env python3
"""Per-language release helper for traust-sdk.

Each SDK under go/, python/, typescript/ has its own VERSION and git tag
prefix (go/vX.Y.Z, python/vX.Y.Z, ts/vX.Y.Z). Tags are independent.
"""

from __future__ import annotations

import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path

_CI_DIR = Path(__file__).resolve().parent / "ci"
if str(_CI_DIR) not in sys.path:
    sys.path.insert(0, str(_CI_DIR))

from languages import (  # noqa: E402
    Language,
    RegistryError,
    language_by_id,
    load_registry,
    read_version,
    semver_tuple,
)

SEMVER_PARTS = frozenset({"major", "minor", "patch"})


class ReleaseError(Exception):
    pass


def bump_version(current: str, part: str) -> str:
    import re

    semver = re.compile(r"^\d+\.\d+\.\d+$")
    if part in SEMVER_PARTS:
        major, minor, patch = map(int, current.split("."))
        if part == "major":
            return f"{major + 1}.0.0"
        if part == "minor":
            return f"{major}.{minor + 1}.0"
        return f"{major}.{minor}.{patch + 1}"
    if semver.match(part):
        return part
    raise ReleaseError(f"bump must be major, minor, patch, or X.Y.Z; got {part!r}")


@dataclass(frozen=True)
class GitOpts:
    force: bool = False
    allow_dirty: bool = False


class Git:
    def __init__(self, cwd: Path, opts: GitOpts | None = None) -> None:
        self.cwd = cwd
        self.opts = opts or GitOpts()

    def _run(
        self,
        *args: str,
        check: bool = True,
        capture: bool = False,
    ) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            args, cwd=self.cwd, check=check, text=True, capture_output=capture
        )

    def status(self) -> str:
        return self._run("git", "status", "--porcelain", capture=True).stdout.strip()

    def branch(self) -> str:
        return self._run(
            "git", "rev-parse", "--abbrev-ref", "HEAD", capture=True
        ).stdout.strip()

    def head(self) -> str:
        return self._run(
            "git", "rev-parse", "--short", "HEAD", capture=True
        ).stdout.strip()

    def tags(self, pattern: str) -> list[str]:
        out = self._run(
            "git", "tag", "--list", pattern, "--sort=-v:refname", capture=True
        ).stdout.strip()
        return out.splitlines() if out else []

    def tag_exists(self, tag: str) -> bool:
        return (
            self._run(
                "git",
                "rev-parse",
                "--verify",
                f"refs/tags/{tag}",
                check=False,
                capture=True,
            ).returncode
            == 0
        )

    def create_tag(self, tag: str, message: str) -> None:
        if self.opts.force and self.tag_exists(tag):
            self._run("git", "tag", "-d", tag)
        self._run("git", "tag", "-a", tag, "-m", message)

    def push_refs(self, *refs: str) -> None:
        cmd = ["git", "push"]
        if self.opts.force:
            cmd.append("--force")
        cmd.append("origin")
        for ref in refs:
            self._run(*cmd, ref)

    @property
    def dirty(self) -> bool:
        return bool(self.status())

    def require_clean(self) -> None:
        if not self.opts.allow_dirty and self.dirty:
            raise ReleaseError("working tree is not clean")


class LanguageRelease:
    def __init__(self, root: Path, lang: Language, git: Git) -> None:
        self.root = root
        self.lang = lang
        self.git = git

    def version(self) -> str:
        return read_version(self.lang, self.root)

    def set_version(self, version: str) -> None:
        self._require_tag_missing(version)
        path = self.lang.version_path(self.root)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(f"{version}\n", encoding="utf-8")

    def tag_name(self, version: str | None = None) -> str:
        return self.lang.tag_name(version or self.version())

    def latest_tag_version(self) -> str | None:
        prefix = f"{self.lang.tag_prefix}/v"
        tags = self.git.tags(f"{self.lang.tag_prefix}/v*")
        if not tags:
            return None
        tag = tags[0]
        if tag.startswith(prefix):
            return tag[len(prefix) :]
        return None

    def status_line(self) -> str:
        version = self.version()
        tag = self.tag_name(version)
        tagged = self.git.tag_exists(tag)
        latest = self.git.tags(f"{self.lang.tag_prefix}/v*")
        latest_tag = latest[0] if latest else "none"
        return (
            f"{self.lang.id}: {version} ({'tagged' if tagged else 'not tagged'}, "
            f"latest tag {latest_tag})"
        )

    def check(self) -> None:
        self.git.require_clean()
        self._require_tag_missing(self.version())

    def tag(self, message: str | None = None) -> str:
        version = self.version()
        tag = self.tag_name(version)
        self._require_tag_missing(version)
        self.git.create_tag(tag, message or f"{self.lang.name} SDK {version}")
        return tag

    def push_tag(self) -> None:
        tag = self.tag_name()
        if not self.git.tag_exists(tag):
            raise ReleaseError(f"tag {tag} does not exist locally")
        self.git.push_refs(tag)

    def release_if_ready(
        self, *, message: str | None = None, push: bool = False
    ) -> str | None:
        current = self.version()
        latest = self.latest_tag_version()
        tag = self.tag_name(current)

        if self.git.tag_exists(tag):
            print(f"{self.lang.id}: {tag} already exists")
            return None

        if latest is not None and semver_tuple(current) == semver_tuple(latest):
            print(
                f"{self.lang.id}: no pending release (VERSION {current} matches latest tag)"
            )
            return None

        if latest is not None and semver_tuple(current) < semver_tuple(latest):
            raise ReleaseError(
                f"{self.lang.id}: VERSION {current} is behind latest tag {self.lang.tag_name(latest)}"
            )

        self.git.require_clean()
        self.tag(message or f"{self.lang.name} SDK {current}")
        print(f"{self.lang.id}: release ready at {tag}")
        if push:
            self.push_tag()
            print(f"{self.lang.id}: pushed {tag}")
        return tag

    def _require_tag_missing(self, version: str) -> None:
        tag = self.tag_name(version)
        if self.git.tag_exists(tag) and not self.git.opts.force:
            raise ReleaseError(f"tag {tag} already exists")


class Workspace:
    def __init__(self, root: Path, opts: GitOpts) -> None:
        self.root = root
        self.git = Git(root, opts)
        self.registry = load_registry()

    def releasable_languages(self) -> list[Language]:
        return [lang for lang in self.registry if lang.is_releasable(self.root)]

    def release_for(self, lang_id: str) -> LanguageRelease:
        return LanguageRelease(
            self.root, language_by_id(lang_id, self.registry), self.git
        )

    def print_status(self) -> None:
        print(
            f"traust-sdk @ {self.git.branch()}@{self.git.head()} "
            f"({'dirty' if self.git.dirty else 'clean'})"
        )
        for lang in self.registry:
            if lang.is_releasable(self.root):
                print(f"  {LanguageRelease(self.root, lang, self.git).status_line()}")
            elif lang.is_present(self.root):
                print(f"  {lang.id}: present (no {lang.dir}/VERSION yet)")

    def release_if_ready_all(
        self, *, message: str | None = None, push: bool = False
    ) -> list[str]:
        tags: list[str] = []
        for lang in self.releasable_languages():
            release = LanguageRelease(self.root, lang, self.git)
            if tag := release.release_if_ready(message=message, push=push):
                tags.append(tag)
        if not tags:
            print("traust-sdk: no language releases pending")
        return tags


def main(argv: list[str] | None = None) -> int:
    import argparse

    flags = argparse.ArgumentParser(add_help=False)
    flags.add_argument("-m", "--message")
    flags.add_argument("--allow-dirty", action="store_true")
    flags.add_argument("--push", action="store_true")
    flags.add_argument("--force", action="store_true")
    flags.add_argument("--lang", action="append", dest="langs")
    flags.add_argument("--all", action="store_true")

    p = argparse.ArgumentParser(
        description="Per-language release helper", parents=[flags]
    )
    sub = p.add_subparsers(dest="cmd", required=True)

    sub.add_parser("status", parents=[flags])
    bump = sub.add_parser("bump", parents=[flags])
    bump.add_argument("lang")
    bump.add_argument("part", help="major, minor, patch, or X.Y.Z")
    check = sub.add_parser("check", parents=[flags])
    check.add_argument("lang")
    tag = sub.add_parser("tag", parents=[flags])
    tag.add_argument("lang")
    push = sub.add_parser("push", parents=[flags])
    push.add_argument("lang")
    rif = sub.add_parser("release-if-ready", parents=[flags])

    args = p.parse_args(argv)
    opts = GitOpts(force=args.force, allow_dirty=args.allow_dirty)
    ws = Workspace(Path.cwd(), opts)

    try:
        match args.cmd:
            case "status":
                ws.print_status()
            case "bump":
                release = ws.release_for(args.lang)
                current = release.version()
                new = bump_version(current, args.part)
                if new == current:
                    print(f"{args.lang}: already at {current}")
                else:
                    release.set_version(new)
                    print(f"{args.lang}: bumped {current} -> {new}")
            case "check":
                release = ws.release_for(args.lang)
                release.check()
                print(f"{args.lang}: ok ({release.tag_name()})")
            case "tag":
                release = ws.release_for(args.lang)
                ws.git.require_clean()
                print(f"{args.lang}: created {release.tag(args.message)}")
            case "push":
                ws.release_for(args.lang).push_tag()
                print(f"{args.lang}: pushed")
            case "release-if-ready":
                if args.all or not args.langs:
                    ws.release_if_ready_all(message=args.message, push=args.push)
                else:
                    for lang_id in args.langs:
                        ws.release_for(lang_id).release_if_ready(
                            message=args.message, push=args.push
                        )
    except (ReleaseError, RegistryError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
