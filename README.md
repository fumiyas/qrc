QR code generator for text terminals
======================================================================

  * Copyright (C) 2014 SATOH Fumiyasu @ OSS Technology Corp., Japan
  * License: MIT License
  * Development home: <https://github.com/fumiyas/qrc>
  * Author's home: <https://fumiyas.github.io/>

What's this?
---------------------------------------------------------------------

This program generates QR codes in
[ASCII art](http://en.wikipedia.org/wiki/ASCII_art) or
[Sixel](http://en.wikipedia.org/wiki/Sixel) format for
text terminals, e.g., console, xterm,
[mlterm](http://sourceforge.net/projects/mlterm/),
Windows command prompt and so on.

Usage
---------------------------------------------------------------------

```console
$ qrc --help
Usage: qrc [OPTIONS] [TEXT]

Options:
  -h, --help
    Show this help message
  -i, --invert
    Invert color

Text examples:
  http://www.example.jp/
  MAILTO:foobar@example.jp
  WIFI:S:myssid;T:WPA;P:pass123;;
$ qrc https://fumiyas.github.io/
...
$ qrc 'WIFI:S:Our-ssid;T:WPA;P:secret;;'
...
```

![optimized](qrc-demo.png)

Download
---------------------------------------------------------------------

Binary files are here for Linux, Mac OS X and Windows:

  * https://github.com/fumiyas/qrc/releases

Build from source codes
---------------------------------------------------------------------

If you have Go language environment, try the following:

```console
$ go get github.com/fumiyas/qrc/cmd/qrc
```

TODO
----------------------------------------------------------------------

  * Add the following options:
    * `--format <aa|sixel>`
    * `--aa-color-scheme <ansi|windows>`
    * `--foreground-color R:G:B`
    * `--background-color R:G:B`
    * `--margin-color R:G:B`
    * `--margin-size N`
    * `--input-encoding E`

