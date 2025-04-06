# devbox

## bash completions

Completions path for devbox installed tools:

`ll $HOME/.local/share/devbox/global/default/.devbox/nix/profile/default/share/bash-completion/completions/`

`ll $HOME/.nix-profile/`

`echo $XDG_DATA_DIRS` shows `${HOME}/.nix-profile/share` and some other pats.

Symlink the share profile:

`ln -s ${HOME}/.local/share/devbox/global/default/.devbox/nix/profile/default/share ${HOME}/.nix-profile/share`

And load the git-promt script:

`source ~/.nix-profile/share/bash-completion/completions/git-prompt.sh`

https://github.com/git/git/blob/master/contrib/completion/git-prompt.sh
https://gitscripts.com/bash-git-prompt

## cockroach client

```sh
$PWD/cockroach sql --certs-dir=certs

CERTS_DIR="${PWD}/certs"

export DATABASE_URL="postgresql://root@localhost:26257/garagesales?sslmode=verify-full&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"

$PWD/cockroach convert-url --url $DATABASE_URL
```
