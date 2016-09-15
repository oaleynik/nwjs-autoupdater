const {spawn} = nw.require('child_process');
const {createHash} = nw.require('crypto');
const {createReadStream, createWriteStream, stat, unlink} = nw.require('fs');
const {tmpdir} = nw.require('os');
const {dirname, resolve} = nw.require('path');
const {execPath, platform} = process;
const semver = nw.require('semver');
const shell = nw.require('shelljs');

const SYSTEM_TEMP_DIR = tmpdir();
const UPDATER_TEMP_DIR = resolve(SYSTEM_TEMP_DIR, 'my-app-updater');
const UPDATES_DIR = resolve(UPDATER_TEMP_DIR, 'updates');
const UPDATER_BIN = resolve(UPDATER_TEMP_DIR, /^win/.test(platform) ? 'updater.exe' : 'updater');

shell.mkdir('-p', UPDATES_DIR);

const {manifest: appManifest} = nw.App;
const {manifestUrl: remoteManifestUrl} = appManifest;

const {
  appInstDir,
  bundledUpdaterPath,
} = resolvePaths(execPath, platform);

run('index.html', appManifest.window)
  .then(() => {
    return fetchManifest(remoteManifestUrl)
      .then(remoteManifest => {
        const {version: currentVersion} = appManifest;
        const {version: latestVersion, [platform]: bundle} = remoteManifest;

        if (semver.gt(latestVersion, currentVersion)) {
          const bundlePath = resolve(UPDATES_DIR, hashString(latestVersion));

          return fetchUpdate(bundle, bundlePath)
            .then(notifyUser)
            .then(result => {
              if (result) {
                startUpdate(bundlePath);
              }
            });
        }
      });
  })
  .catch(() => {});

function resolvePaths (execPath, platform) {
  let appDir;
  let appInstDir;
  let appExec;
  let bundledUpdaterPath;

  if (platform === 'darwin') {
    appDir = resolve(execPath, '../../../../../../../');
    appInstDir = dirname(appDir);
    appExec = appDir;
    bundledUpdaterPath = resolve(appDir, 'Contents', 'Resources', 'updater');
  } else if (platform === 'win32') {
    appDir = dirname(execPath);
    appInstDir = appDir;
    appExec = resolve(appDir, 'MyApp.exe');
    bundledUpdaterPath = resolve(appDir, 'updater.exe');
  } else {
    appDir = dirname(execPath);
    appInstDir = appDir;
    appExec = resolve(appDir, 'MyApp');
    bundledUpdaterPath = resolve(appDir, 'updater');
  }

  return {
    appDir,
    appInstDir,
    appExec,
    bundledUpdaterPath,
  };
}

function hashString (value, algorithm = 'sha256', inputEncoding = 'latin1', digestEncoding = 'hex') {
  return createHash(algorithm).update(value, inputEncoding).digest(digestEncoding);
}

function startUpdate (bundlePath) {
  shell.cp(bundledUpdaterPath, UPDATER_BIN);
  shell.chmod(755 & ~process.umask(), UPDATER_BIN);

  spawn(UPDATER_BIN, [
    '--bundle', bundlePath,
    '--inst-dir', appInstDir,
  ], {
    cwd: dirname(UPDATER_BIN),
    detached: true,
    stdio: 'ignore',
  }).unref();

  nw.App.quit();
}

function fetchUpdate ({url, sha256}, dest) {
  return fileExists(dest)
    .then(exists => {
      if (exists) {
        return checkSHA(dest, sha256)
          .then(() => Promise.resolve(dest))
          .catch(err => {
            if (/^SHA256 mismatch/.test(err.message)) {
              return removeFile(dest)
                .then(() => downloadFile(url, dest, sha256));
            }

            return Promise.reject(err);
          });
      }

      return downloadFile(url, dest, sha256);
    });
}

function fetchManifest (url) {
  return new Promise((resolve, reject) => {
    const http = /^https/.test(url) ? nw.require('https') : nw.require('http');

    http.get(url, res => {
      if (res.statusCode !== 200) {
        return reject(new Error(res.statusMessage));
      }

      const buffer = [];

      res.on('data', chunk => buffer.push(chunk));
      res.on('end', () => {
        const raw = Buffer.concat(buffer).toString();
        const manifest = JSON.parse(raw);

        resolve(manifest);
      });
    }).on('error', err => reject(err));
  });
}

function downloadFile (source, dest, sha256 = false) {
  return new Promise((resolve, reject) => {
    const ws = createWriteStream(dest);
    const http = /^https/.test(source) ? nw.require('https') : nw.require('http');

    http.get(source, res => {
      if (res.statusCode !== 200) {
        return reject(new Error(res.statusMessage));
      }

      res.pipe(ws).on('finish', () => {
        if (!sha256) {
          return resolve(dest);
        }

        checkSHA(dest, sha256)
          .then(() => resolve(dest))
          .catch(err => reject(err));
      });
    }).on('error', err => reject(err));
  });
}

function checkSHA (filepath, sha256) {
  return new Promise((resolve, reject) => {
    const rs = createReadStream(filepath);
    const hash = createHash('sha256');

    rs.pipe(hash).on('data', digest => {
      const digestHex = digest.toString('hex');

      if (digestHex !== sha256) {
        return reject(new Error(`SHA256 mismatch: ${sha256} !== ${digestHex}`));
      }

      resolve();
    });
  });
}

function notifyUser () {
  return new Promise(resolve => {
    const options = {
      icon: 'icons/update.png',
      body: 'Click here to install',
    };

    const notification = new Notification('A new update is available!', options);

    notification.onclick = () => {
      notification.close();

      resolve(true);
    };

    notification.onclose = () => resolve(false);
  });
}

function run (entryFile, windowParams) {
  return new Promise(resolve => {
    nw.Window.open(entryFile, windowParams, win => resolve(win));
  });
}

function removeFile (filepath) {
  return new Promise((resolve, reject) => {
    unlink(filepath, err => {
      if (err) {
        return reject(err);
      }

      resolve();
    });
  });
}

function fileExists (bundledUpdaterPath) {
  return new Promise(resolve => {
    stat(bundledUpdaterPath, (err, stats) => {
      if (err) {
        if (err.code === 'ENOENT') {
          return resolve(false);
        }

        throw err;
      }

      if (stats.isFile()) {
        return resolve(true);
      }

      resolve(false);
    });
  });
}
