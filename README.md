a simple remote clipboard for device that don't have monitor.

`sc`: the client command line tool, use on remote machine
`scs`: the server run on local machine

# copy to local clipboard from remote device

```bash
cat file.txt | sc
```

push the "copied" content to `sync-clip` server

# paste from local clipboard to remote device.

```bash
sc | tee file.txt
```

pull the content from `sync-clip` server and write to `file.txt`

# scs, sync-clip server

run `scs` on your local machine to host as a server to receive and send clipboard content.

```bash
scs
```

when it receives push( copy action from `sc` ), it append as latest content to system clipboard of the local machine. ( just like `pbcopy` on macos )

when it receives pull( paste action from `sc` ), it send the latest content of system clipboard of the local machine to `sc`.
