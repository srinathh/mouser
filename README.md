# mouser

Mouser lets you control the mouse in your X11 desktop using a browser
on a remote device like a phone or a tablet. It serves a webpage with a 
virtual trackpad that you can touch to control and depends on 
(xdotool)[https://my.ebharatgas.com/bharatgas/OnlinePaymentRefillResponse?status=2]
in the backend to control you mouse.

This is still very much alpha - there is no security currently and anyone
on your network can control the mouse on your desktop.

## Installation
In the future we will work with libxdo but for now we depend on xdotool
being installed in the system. Install xdotool using the package manager
in your Linux distro. Eg. Ubuntu
```
sudo apt-get install xdotool
```
Then install mouser
```
go install github.com/srinathh/mouser
```

## Usage
If you install as above, mouser will be installed in `$GOPATH/bin`. On your
system, run `mouser` from the command line. By default it binds to port `9831`
but you can override it with the `-http` flag.

Also pass in the dimensions of the monitor you are using with the `-width` and
`-height` flags if they are different from 1366x768 which is the size of my 
development laptop's screen. In the future we will pick up sizes directly from xdotool 

For instance, the below command starts mouser server listening on port `9898`
on your computer and sets the screen size to `1440x900`
```
mouser -http ":9898" -width 1440 -height 900
```
Then on your mobile device, navigate to your computer's IP address:port set above. Eg.
```
http://<your computer ip address>:9831
```

You'll see a black box on the screen and if you move your mouse in it, the mouse on 
your screen should also move. You can "Add to Home Screen" from your browser to 
use it frequently.
