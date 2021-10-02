import build_core

def linux():
    build_core.build("linux", "amd64", "westpunk")

if __name__ == "__main__":
    # linux() # THIS OVERWRITES THE MACOS EXECUTABLE SO SHOULD ONLY BE CALLED AS A PART OF BUILD.PY
    pass