## Capture ARGS from git in the script

Lefthook passes Git arguments to your commands and scripts.

```
├── .lefthook
│   └── prepare-commit-msg
│       └── message.sh
└── lefthook.yml
```

```yml
# lefthook.yml

prepare-commit-msg:
  jobs:
    - script: "message.sh"
      runner: bash
    - run: echo "Git args: {1} {2} {3}"
```

```bash
# .lefthook/prepare-commit-msg/message.sh

# Arguments get passed from Git

COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

# ...
```

