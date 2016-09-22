*Tiny golang app, which can be bundled with your NWJS application (actually, any application) and used for "autoupdates." I really hope that we will have something like this out-of-the box in NWJS!*

### Problem 

NWJS is amazing platform for building desktop applications using web technologies. However, it is missing one important feature - an ability to seamlessly deliver updates for users.
There were several attempts to solve this problem. For example - https://github.com/edjafarov/node-webkit-updater. But it does have issues when updater itself needs to be updated or NWJS platform needs to be updated (https://github.com/nwjs/nw.js/issues/233).  

### Solution

This tiny golang application (when built it is just ~2MB) can be bundled with your NWJS application and then used to unpack updates.
To update target application updater needs to know two things - where zip archive with the new version is located and where is the app's executable to restart application after update. These can be passed to updater via command line arguments `--bundle` and `--inst-dir`, where `--bundle` is the path to the zip archive with the new app version and `--inst-dir` is the path to app's executable.  

### Build

1) You need to have `golang` installed and properely configured ;)  
2) Ensure that you have proper information in the manifest file describing your application. Make sure that requested execution level is set to `asInvoker`. **Without it your main app will not be able to run autoupdater**.  
3) Install `go get github.com/akavel/rsrc`  
4) Compile your manifest file into `.syso` file using `rsrc` tool. For example:
```
    rsrc -manifest updater.exe.manifest [-ico FILE.ico[,FILE2.ico...]] -o updater.syso
```
5) Build autoupdater binaries for using one of the following commands (for OSX, Win32 and Win64 respectively):  
```
    GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o updater 
    GOOS=windows GOARCH=386 go build -ldflags "-s -w -H=windowsgui" -o updater.exe
    GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -H=windowsgui" -o updater.exe
```

> Full list of the platforms and archetectures, which is possible to cross-compile to can be found here https://github.com/golang/go/blob/master/src/go/build/syslist.go

### Example

Let's consider the small example to see how it can be used.

- Start application from the index.js (instead of directly pointing it to the .html file) ([Example](https://github.com/oaleynik/nwjs-tiny-updater/blob/master/examples/index.js))
- Open main window (so user will not have to wait)
- Create temporary folder in the system's temp directory (let's call it T)
- Create `updates` folder in the T folder.
- Download manifest file from the remote location. It has the format, which is similar to the one used in https://github.com/edjafarov/updater. For example:
```json
{
  "version": "1.1.0",
  "createdAt": "2016-09-15T15:02:02.365Z",
  "win32": {
    "url": "https://example.com/windows-x32-v1.1.0.zip",
    "sha256": "a5db62bbe2c382534162d921ae6319ae83384256ab7f40e928a323603653e22b"
  },
  "darwin": {
    "url": "https://example.com/darwin-x64-v1.1.0.zip",
    "sha256": "e9efa23c40cf17a1d0c77227c962b5659c099959ec199782e03f162dcbbdae19"
  }
}
```
- Compare local app's version with the one from manifest file
- If app has the latest version from manifest - do nothing
- If app is outdated - download zip archive to the `updates` folder created before. When it is downloaded - show notification for user using `new Notification` or `chrome.notifications.create`. If user clicks on notification - proceed.
- Copy `bundled updater executable` to the folder T (don't move it because temp folders sometimes can be, you know, - cleaned :)
- Run copy of the updater executable using `spawn` in detached mode and close application. Pass path to the downloaded archive and path to the main app executable to updater using command line arguments.
- When updater starts it unpacks zip archive to the temp directory. Than it creates backup of the old app by renaming parent directory to .bak (on macOS the Application.app will be renamed itself)
- Move unpacked files to the new location. If all went well - delete bak, delete temp files, delete zip archive and start main application.

Due to the fact that we copy and run updater from the temp directory - any files in application can be updated!

### Windows updates

Autoupdates on windows were really painfull. All those "Run as Administrator" and issues with replacing executable, which is running :( But, this tiny updater solves this issue too. In the root of the repo you can find `updater.exe.manifest`, which defines metadata of the binary and one important (for Windows) thing - permissions level. The line `<requestedExecutionLevel level="asInvoker" uiAccess="false"/>` will allow your app to run updater on windows without admin rights. To embed this information in your updater's binary you need to use `https://github.com/akavel/rsrc` - amazing tool for embedding binary resources in Go programs. Also you can pass the path to the .ico file and updater's binary will get a nice icon ;)

### Why golang?

Because it is simple and more importantly - it can be compiled from any platform for any platform.

### Credits
I'm not very proficient in the golang (yet) so I've used and slightly modified several packages from the go community (all credits are going to them):

https://github.com/skratchdot/open-golang - Open a file, directory, or URI using the OS's default application for that object type. Optionally, you can specify an application to use.  
https://github.com/ivaxer/go-xattr - Extended attribute support for Go #golang  
And very simple zip unpacker (can't find the source)  
