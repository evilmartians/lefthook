from setuptools import setup, find_packages

setup(
    name='lefthook',
    version='1.7.18',
    author='Evil Martians',
    author_email='lefthook@evilmartians.com',
    url='https://github.com/evilmartians/lefthook',
    description='A single dependency-free binary to manage all your git hooks that works with any language in any environment, and in all common team workflows',
    packages=find_packages(),
    entry_points={
        'console_scripts': [
            'lefthook=lefthook.main:main'
        ],
    },
    package_data={
        'lefthook':[
            'bin/lefthook-linux-x86_64/lefthook',
            'bin/lefthook-linux-arm64/lefthook',
            'bin/lefthook-freebsd-x86_64/lefthook',
            'bin/lefthook-freebsd-arm64/lefthook',
            'bin/lefthook-openbsd-x86_64/lefthook',
            'bin/lefthook-openbsd-arm64/lefthook',
            'bin/lefthook-windows-x86_64/lefthook.exe',
            'bin/lefthook-windows-arm64/lefthook.exe',
            'bin/lefthook-darwin-x86_64/lefthook',
            'bin/lefthook-darwin-arm64/lefthook',
        ]
    },
    classifiers=[
        'License :: OSI Approved :: MIT License',
        'Operating System :: OS Independent',
        'Topic :: Software Development :: Version Control :: Git'
    ],
    python_requires='>=3.6',
)
