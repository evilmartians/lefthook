import atexit
import os
import platform
import shutil
import sys
import tempfile
from pathlib import Path

from hatchling.builders.hooks.plugin.interface import BuildHookInterface


PLATFORM_MAPPING = {
    'linux': 'linux',
    'linux2': 'linux',
    'darwin': 'darwin',
    'win32': 'windows',
    'windows': 'windows',
    'freebsd': 'freebsd',
    'openbsd': 'openbsd',
}

ARCH_MAPPING = {
    'x86_64': 'x86_64',
    'amd64': 'x86_64',
    'arm64': 'arm64',
    'aarch64': 'arm64',
}


PEP425_TAGS = {
    ("linux", "x86_64"): "py3-none-manylinux_2_17_x86_64",
    ("linux", "arm64"): "py3-none-manylinux_2_17_aarch64",
    ("darwin", "x86_64"): "py3-none-macosx_10_15_x86_64",
    ("darwin", "arm64"): "py3-none-macosx_11_0_arm64",
    ("windows", "x86_64"): "py3-none-win_amd64",
    ("windows", "arm64"): "py3-none-win_arm64",
}


def normalize_platform(value: str) -> str:
    if not value:
        return value
    return PLATFORM_MAPPING.get(value.lower(), value.lower())


def normalize_arch(value: str) -> str:
    if not value:
        return value
    return ARCH_MAPPING.get(value.lower(), value.lower())


def get_platform_info():
    target_platform = os.environ.get('LEFTHOOK_TARGET_PLATFORM')
    target_arch = os.environ.get('LEFTHOOK_TARGET_ARCH')

    if target_platform and target_arch:
        normalized_platform = normalize_platform(target_platform)
        normalized_arch = normalize_arch(target_arch)
        print(f"[HOOK] Using target: {normalized_platform}-{normalized_arch}")
        return normalized_platform, normalized_arch

    system = normalize_platform(sys.platform) or normalize_platform(platform.system())
    machine = normalize_arch(platform.machine())
    result = system, machine
    print(f"[HOOK] Auto-detected: {result[0]}-{result[1]}")
    return result


class CustomBuildHook(BuildHookInterface):
    PLUGIN_NAME = "custom"

    def __init__(self, *args, **kwargs) -> None:
        super().__init__(*args, **kwargs)
        self.target_platform = None
        self.target_arch = None
        self._temp_dir = None
        self._moved_entries = []
        self._restore_registered = False

    def initialize(self, version, build_data):
        target_platform, target_arch = get_platform_info()
        self.target_platform = target_platform
        self.target_arch = target_arch

        tag = PEP425_TAGS.get((target_platform, target_arch))
        if tag:
            build_data["tag"] = tag
            self._prune_binaries()
            if not self._restore_registered:
                atexit.register(self._restore_binaries)
                self._restore_registered = True
            print(f"[HOOK] Building platform wheel {tag}")
        else:
            print(
                "[HOOK] No PEP425 tag for "
                f"{target_platform}-{target_arch}; building universal wheel."
            )

        print(f"[HOOK] Initialized for {target_platform}-{target_arch}")

    def finalize(self, version, build_data, artifact_path) -> None:
        print(f"[HOOK] Built artifact: {artifact_path}")
        self._restore_binaries()

    def _prune_binaries(self):
        if not self.target_platform or not self.target_arch:
            raise RuntimeError("Target platform is not set before pruning binaries.")

        bin_dir = Path(self.root) / "lefthook" / "bin"
        if not bin_dir.is_dir():
            raise RuntimeError(f"Bin directory not found: {bin_dir}")

        target_dir_name = f"lefthook-{self.target_platform}-{self.target_arch}"
        target_dir = bin_dir / target_dir_name
        if not target_dir.exists():
            available = ", ".join(sorted(p.name for p in bin_dir.iterdir() if p.is_dir()))
            raise FileNotFoundError(
                f"Binary folder '{target_dir_name}' is missing. Available: {available or 'none'}"
            )

        binaries = list(target_dir.glob("lefthook*"))
        if not binaries:
            raise FileNotFoundError(
                f"No lefthook binary found under {target_dir}."
            )

        self._temp_dir = Path(tempfile.mkdtemp(prefix="lefthook-bin-backup-"))
        preserved = {target_dir_name, ".keep"}

        for entry in bin_dir.iterdir():
            if entry.name in preserved:
                continue
            destination = self._temp_dir / entry.name
            shutil.move(str(entry), str(destination))
            self._moved_entries.append((destination, entry))

        print(f"[HOOK] Shipped binaries: {target_dir_name}")

    def _restore_binaries(self):
        while self._moved_entries:
            backup_path, original_path = self._moved_entries.pop()
            if backup_path.exists():
                shutil.move(str(backup_path), str(original_path))
        if self._temp_dir and self._temp_dir.exists():
            shutil.rmtree(self._temp_dir, ignore_errors=True)
        self._temp_dir = None
