#!/usr/bin/env python3
"""Go SDK CI — toolchain setup, lint, test, and security gates."""

from __future__ import annotations

import argparse
import os
import re
import shutil
import subprocess
import sys
from collections.abc import Callable, Sequence
from dataclasses import dataclass
from pathlib import Path

# --- pins (bump deliberately) ---

DEFAULT_TOOLCHAIN = "go1.26.6"
GOVULNCHECK_VERSION = "v1.1.4"
GOSEC_VERSION = "v2.22.2"

GO_INSTALL_PACKAGES = (
    f"golang.org/x/vuln/cmd/govulncheck@{GOVULNCHECK_VERSION}",
    f"github.com/securego/gosec/v2/cmd/gosec@{GOSEC_VERSION}",
)

SECURITY_BINARIES = ("govulncheck", "gosec")

# --- declarative command tables ---

SETUP_CMD = ("go", "version")

LINT_CMDS: tuple[tuple[str, ...], ...] = (
    ("go", "vet", "./..."),
    ("make", "fmt-check"),
)

TEST_CMDS: tuple[tuple[str, ...], ...] = (("go", "test", "./..."),)

SECURITY_CMDS: tuple[tuple[str, ...], ...] = (
    ("govulncheck", "./..."),
    (
        "gosec",
        "-exclude-generated",
        "-exclude-dir=internal",
        "-fmt=text",
        "./...",
    ),
)

TOOLCHAIN_RE = re.compile(r"^toolchain\s+(\S+)", re.MULTILINE)


@dataclass(frozen=True)
class LabeledCmd:
    label: str
    argv: tuple[str, ...]


SECURITY_STEPS = tuple(
    LabeledCmd(label=argv[0], argv=argv) for argv in SECURITY_CMDS
)


def repo_root() -> Path:
    return Path(os.environ.get("CI_PROJECT_DIR", Path(__file__).resolve().parents[3]))


def read_toolchain(go_mod: Path) -> str:
    if not go_mod.is_file():
        return DEFAULT_TOOLCHAIN
    match = TOOLCHAIN_RE.search(go_mod.read_text(encoding="utf-8"))
    return match.group(1) if match else DEFAULT_TOOLCHAIN


def gopath_bin(env: dict[str, str]) -> str:
    out = subprocess.run(
        ("go", "env", "GOPATH"),
        capture_output=True,
        text=True,
        check=True,
        env=env,
    )
    return str(Path(out.stdout.strip()) / "bin")


def build_env(root: Path) -> dict[str, str]:
    go_dir = root / "go"
    env = os.environ.copy()
    env["GOTOOLCHAIN"] = read_toolchain(go_dir / "go.mod")
    env["GOBIN"] = env.get("GOBIN", gopath_bin(env))
    env["PATH"] = f"{env['GOBIN']}{os.pathsep}{env.get('PATH', '')}"
    return env


def run(
    argv: Sequence[str],
    *,
    cwd: Path,
    env: dict[str, str],
) -> None:
    try:
        subprocess.run(argv, cwd=cwd, env=env, check=True)
    except subprocess.CalledProcessError as exc:
        raise SystemExit(exc.returncode) from exc


def tools_missing(env: dict[str, str]) -> bool:
    return any(shutil.which(name, path=env["PATH"]) is None for name in SECURITY_BINARIES)


@dataclass(frozen=True)
class GoCI:
    root: Path

    @property
    def go_dir(self) -> Path:
        return self.root / "go"

    def env(self) -> dict[str, str]:
        return build_env(self.root)

    def setup(self) -> None:
        run(SETUP_CMD, cwd=self.go_dir, env=self.env())

    def run_cmds(self, cmds: Sequence[tuple[str, ...]]) -> None:
        env = self.env()
        for argv in cmds:
            run(argv, cwd=self.go_dir, env=env)

    def install_tools(self) -> None:
        self.setup()
        env = self.env()
        for package in GO_INSTALL_PACKAGES:
            run(("go", "install", package), cwd=self.go_dir, env=env)

    def lint(self) -> None:
        self.setup()
        self.run_cmds(LINT_CMDS)

    def test(self) -> None:
        self.setup()
        self.run_cmds(TEST_CMDS)

    def security(self) -> None:
        env = self.env()
        if tools_missing(env):
            self.install_tools()
            env = self.env()

        for step in SECURITY_STEPS:
            print(f"==> {step.label}")
            run(step.argv, cwd=self.go_dir, env=env)


CommandHandler = Callable[[GoCI], None]

COMMANDS: dict[str, CommandHandler] = {
    "setup": GoCI.setup,
    "lint": GoCI.lint,
    "test": GoCI.test,
    "tools": GoCI.install_tools,
    "security": GoCI.security,
}


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Go SDK CI runner")
    sub = parser.add_subparsers(dest="cmd", required=True)
    for name in COMMANDS:
        sub.add_parser(name)

    args = parser.parse_args(argv)
    COMMANDS[args.cmd](GoCI(repo_root()))
    return 0


if __name__ == "__main__":
    sys.exit(main())
