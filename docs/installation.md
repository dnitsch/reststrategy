# Installation

Major platform binaries [here](https://github.com/dnitsch/reststrategy/releases)

*nix binary

```bash
curl -L https://github.com/dnitsch/reststrategy/releases/latest/download/seeder-linux -o restseeder
```

MacOS binary

```bash
curl -L https://github.com/dnitsch/reststrategy/releases/latest/download/seeder-darwin -o restseeder
```

Windows

```posh
iwr -Uri "https://github.com/dnitsch/reststrategy/releases/latest/download/seeder-windows.exe" -OutFile "restseeder"
```

```bash
chmod +x restseeder
sudo mv restseeder /usr/local/bin
```

>On windows move to a directory which is added to the path (or something similar :shrug: )

Download specific version:

```bash
curl -L https://github.com/dnitsch/reststrategy/releases/download/v0.6.5-pre/seeder-`uname -s` -o restseeder
```
