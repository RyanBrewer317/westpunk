import build_core

def windows():
    build_core.build("windows", "amd64", "westpunk.exe")

if __name__ == "__main__":
    windows()