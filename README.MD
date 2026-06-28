<div align="center">
  <img src="images/logo.svg" alt="logo" width="172" height="172" /> 
  <h2>Resolume Timecode Monitor</h2>
  <p>View the time left for a playing clip on multiple devices.</p>
</div>

| <img alt="client" src="images/demo.png" width="65%"/> | <img alt="server" src="images/demo2.jpg" width="130%"/> |
|:-----------------------------------------------------:|:-------------------------------------------------------:|
|                _The client interface._                |               _The server end of things._               |

## How to use

1. Download the [latest release](https://github.com/calebweixun/resolume-timecode/releases)
2. Open the app and verify that all the information in there is set correctly.
3. Click _Start Server_
4. Open your web browser to the location specified by the application.
5. Start playing a clip in Resolume.
6. Profit.

### macOS: opening the downloaded app

The macOS build is not notarized. If macOS blocks it after you extract the ZIP,
remove the quarantine attribute from this app only, then open it again:

```bash
xattr -dr com.apple.quarantine "/Applications/resolume-timecode.app"
```

Adjust the path if you placed the app somewhere other than `/Applications`.

> Note: The precision should be within a few hundred milliseconds (usually 50ms), if it feels off double check that the clip length 
(in the status bar on the bottom) is correct. If it isn't, just click the reset button.

### How to set the `Path` field

| Clip to use    | Example path                    |
|:---------------|:--------------------------------|
| Selected Clip  | `/composition/selectedclip`     |
| Specific Clip  | `/composition/layers/1/clips/1` |
| Selected Layer | `/composition/selectedlayer`    |
| Specific Layer | `/composition/layers/1`         |

> Note: using a layer will cause the layer's name to be shown on the client instead of the clip name.

## Clock Settings

The web display can be customised from the **Clock Settings** tab in the server GUI — no need to touch the browser.

| Setting | Default | Description |
|:--------|:--------|:------------|
| Show Hours (HH) | Off | Show or hide the hours digits |
| Show Milliseconds | Off | Show or hide the milliseconds digits |
| Show +/− | On | Show or hide the T+/T− sign |
| Clip Name Size | 2.2 vw | Font size of the clip name label |

Changes take effect immediately on all connected clients when you click **Update Settings**.

## OBS / Broadcast use

The web client is designed to fit a **1280×720** browser source out of the box. The timecode fills the full viewport with no scrollbars. Set your OBS browser source to 1280×720 and point it at `http://<server-ip>:8080`.

The display turns **red** when less than 10 seconds remain.

## Motivation

When using Resolume for events where longer clips were being played, I found it annoying that the only way to find out
the remaining time was to ask the guy in front of the laptop. So I decided to fix that.

## How it works

The server polls Resolume's clip position through OSC about every 110ms and refreshes the selected clip's name and duration
about once per second. It calculates the remaining time and broadcasts it to all connected web clients via WebSocket. The
browser then interpolates between updates using `requestAnimationFrame` for a smooth, continuously-running display. Manual
Reset is only needed for troubleshooting.

For the exact algorithm, check [procmessage.go](https://github.com/chabad360/resolume-timecode/blob/master/procmessage.go).

## Known Issues

- Uses around 60-70mb of RAM. Sorry, can't do much about that, at least it's better than an electron app.

## Building

Because this is based on [fyne](https://fyne.io) it requires the use of _cgo_, so make sure you have a C compiler installed.

```bash
git clone https://github.com/chabad360/resolume-timecode
cd resolume-timecode
go build .
```

A `dev.sh` helper script is included for local development:

```bash
./dev.sh          # build + run
./dev.sh send-osc # simulate Resolume OSC output (requires python3)
./dev.sh open     # open browser to web client
./dev.sh clean    # remove build artefacts
```

Releases for macOS (Intel + Apple Silicon) and Windows are built automatically via GitHub Actions when a version tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```
