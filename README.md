# Status bar

This is a statusbar for a **linux** window managers written in **golang**.
Currently it provides these details:

## Screenshot

![Screenshot](https://cloud.githubusercontent.com/assets/132389/11459250/ff0406d4-96da-11e5-8afa-73721233b6f6.png)

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
- **lm_sensors** for cpu temperature detection.
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

    statusbar emails.json | dzen2 -x -820 -y 0 -bg '#073642' -fg '#839496' -ta r -p -fn 'InconsolataSansMono:size=11' -h 19 -w 710

**NOTE:** you may change these properties based on your screen layout and fonts.

![Screenshot](https://cloud.githubusercontent.com/assets/132389/11459253/0ffa8616-96db-11e5-952b-de3d27c5f792.png)

