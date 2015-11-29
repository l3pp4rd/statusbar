# Status bar

This is a statusbar for a **linux** window managers written in **golang**.
Currently it provides these details:

- shows active keyboard layout.
- **network** connection details, wifi or ethernet, upload and download speeds. (**nmcli** is required)
- unread gmail email counts for multiple accounts in given order.
- **cpu** temperature.
- **cpu** load.
- **memory** usage.
- **power** details, AC if on power cable, or remaining **battery** percentage.
- **date** local date and time.

You may need to install:

- **dzen2** is the package used to render the status bar on your X11 screen
- **upower** package to provide battery and AC usage details.
- **networkmanager** which provides __nmcli__ command for network details. Most probably your system uses network manager by default.

## Installation

You must have **go** installed on your system.

This repository is meant to be editable to your own needs, so fork or clone and edit.

If you have gmail accounts, create a config based on **emails.dist.json**.

Installs **go-bindata** on your GOPATH

    make deps

If you run `make` it will build and move binary to **/usr/local/bin/statusbar**.

    make

Now you can run:

    statusbar emails.json | dzen2 -x 50 -y 500 -bg '#073642' -fg '#839496' -ta r -p -fn 'InconsolataSansMono:size=11' -h 20 -w 700

**NOTE:** you may change these properties based on your screen layout and fonts.


