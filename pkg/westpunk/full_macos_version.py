import build_core

def macos():
    build_core.build("darwin", "amd64", "westpunk")

if __name__ == "__main__":
    macos()