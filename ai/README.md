# Setup ML/AI Python packages

Run `go get` in the root directory

## Inside ai/

Setup [gopy](https://github.com/go-python/gopy#installation)

Run

`$ gopy build --output=engine github.com/mrfleap/custom-compression/engine`

Install requirements

`$ pip install -r requirements.txt`

# Running

Use `./start.sh` to run the program from the shell

This exports the engine directory as a LD_LIBRARY_PATH which needs to be set before invoking the python program.

Here's a sample VSCode run configuration:

```json
{
  "name": "test.py",
  "type": "python",
  "request": "launch",
  "program": "test.py",
  "console": "integratedTerminal",
  "env": {
    "LD_LIBRARY_PATH": "${env:LD_LIBRARY_PATH}:${fileDirname}/engine"
  },
  "cwd": "${workspaceFolder}/ai/"
}
```
