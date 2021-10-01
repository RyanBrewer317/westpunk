import sys
import full_windows_version
import full_macos_version

if len(sys.argv) > 1:
    print(sys.argv[0])
    if "windows" in sys.argv:
        full_windows_version.windows()
    if "macos" in sys.argv:
        full_macos_version.macos()
else:
    full_macos_version.macos()
    full_windows_version.windows()