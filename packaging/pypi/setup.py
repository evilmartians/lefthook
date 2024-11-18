from setuptools import setup, find_packages

with open("README.md", "r") as fh:
    long_description = fh.read()

setup(
    name='lefthook',
    version='1.8.4',
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
