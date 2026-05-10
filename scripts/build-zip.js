const { execSync } = require('child_process');
const os = require('os');

if (os.platform() === 'win32') {
  execSync('powershell -command "Compress-Archive -Path bin/bootstrap -DestinationPath bin/api.zip -Force"', { stdio: 'inherit' });
} else {
  execSync('zip -j bin/api.zip bin/bootstrap', { stdio: 'inherit' });
}
