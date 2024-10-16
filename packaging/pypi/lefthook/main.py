import os
import sys
import platform
import subprocess

def main():
    os_name = platform.system().lower()
    ext = os_name == "windows" and ".exe" or ""
    arch = platform.machine()
    subfolder = f"lefthook-{os_name}-{arch}"
    executable = os.path.join(os.path.dirname(__file__), "bin", subfolder, "lefthook"+ext)
    result = subprocess.run([executable] + sys.argv[1:])
    return result.returncode
