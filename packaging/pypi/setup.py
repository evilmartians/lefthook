import os
import sys
import platform
from setuptools import setup, find_packages
from wheel.bdist_wheel import bdist_wheel as _bdist_wheel

with open("README.md", "r") as fh:
    long_description = fh.read()

PLATFORM_MAPPING = {
    'linux': 'linux',
    'darwin': 'darwin',
    'win32': 'windows',
    'freebsd': 'freebsd',
    'openbsd': 'openbsd',
}

ARCH_MAPPING = {
    'x86_64': 'x86_64',
    'AMD64': 'x86_64',
    'amd64': 'x86_64',
    'aarch64': 'arm64',
    'arm64': 'arm64',
}

def get_platform_info():
    """Determine the platform and architecture from environment or system."""
    target_platform, target_architecture = os.environ.get('LEFTHOOK_TARGET_PLATFORM'), os.environ.get('LEFTHOOK_TARGET_ARCH')

    if target_platform and target_architecture:
        return target_platform, target_architecture

    system, machine = platform.system().lower(), platform.machine().lower()
    return PLATFORM_MAPPING.get(sys.platform, system), ARCH_MAPPING.get(machine, machine)

class CustomBdistWheel(_bdist_wheel):
    """Custom bdist_wheel that generates platform-specific wheels."""

    def finalize_options(self):
        """Initialize platform-specific options."""
        super().finalize_options()
        platform, architecture = get_platform_info()
        self.plat_name = f"{platform}_{architecture}"

    def get_tag(self):
        """Override to return platform-specific tag."""
        platform, architecture = get_platform_info()
        return ('py3', 'none', f'{platform}_{architecture}')

def get_package_data():
    """Get only the binary for the current/target platform."""
    platform, architecture = get_platform_info()
    binary_name = 'lefthook'
    if platform == 'windows':
        binary_name = 'lefthook.exe'
    return {
        'lefthook': [f'bin/lefthook-{platform}-{architecture}/{binary_name}']
    }

setup(
    name='lefthook',
    version='2.0.15',
    author='Evil Martians',
    author_email='lefthook@evilmartians.com',
    url='https://github.com/evilmartians/lefthook',
    description='Git hooks manager. Fast, powerful, simple.',
    long_description=long_description,
    long_description_content_type="text/markdown",
    packages=find_packages(),
    entry_points={
        'console_scripts': [
            'lefthook=lefthook.main:main'
        ],
    },
    package_data=get_package_data(),
    license_files=['LICENSE'],
    cmdclass={
        'bdist_wheel': CustomBdistWheel,
    },
    classifiers=[
        'Operating System :: OS Independent',
        'Topic :: Software Development :: Version Control :: Git'
    ],
    python_requires='>=3.6',
)
