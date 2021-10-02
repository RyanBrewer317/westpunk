import sys
import full_windows_version
import full_macos_version
import full_linux_version

if len(sys.argv) > 1:
    if "windows" in sys.argv:
        full_windows_version.windows()
    if "macos" in sys.argv:
        full_macos_version.macos()
    if "linux" in sys.argv:
        full_linux_version.linux()
else:
    full_macos_version.macos()
    full_windows_version.windows()
    full_linux_version.linux()