# Statusbar

This is a statusbar for a **linux** window managers written in **golang**.
Currently it provides these details:

## Screenshot

![Screenshot](https://cloud.githubusercontent.com/assets/132389/11613209/8c0a3260-9c21-11e5-8588-16418956562d.png)

- shows active keyboard layout, using **setxkbmap**.
- **gmail** account unread email counts in configured order.
- **network** connection details, **wifi** or **ethernet**, upload and
  download speeds.
- **cpu** temperature.
- **power** details, **AC** if on power cable, or remaining **battery**
  percentage.
- **cpu** load.
- **memory** utilization percent.
- **date** local date and time, plus one in different timezone in my case.

## Requirements

- **go** in order to compile statusbar.
- **dzen2** is the package used to render the status bar on your X11 screen

## Installation

You must have **go** installed on your system.

This repository is meant to be editable to your own needs, so fork or
clone and edit. Create your statusbar configuration:

    cp statusbar.dist.json statusbar.json

**NOTE:** the arguments for **dzen2** output formatting should be changed
on your needs including **gmail** accounts if available.

If you run `make` it will build and move binary to
**/usr/local/bin/statusbar** and statusbar.json if available, to
**/usr/local/etc/statusbar.json**.

    make

If dependencies were not met, install them. Now you can run statusbar
which takes configuration option json as an argument:

    statusbar statusbar.json > /tmp/statusbar.log 2>&1

**NOTE:** you may change configuration properties based on your screen
layout and fonts. Statusbar logs errors to **stdout** and in case of panic
to **stderr**.

