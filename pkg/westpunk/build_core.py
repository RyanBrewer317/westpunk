from subprocess import run
from zipfile import ZipFile
from pathlib import Path
from shutil import copy
from os import remove, environ
from os.path import join, split, sep
from sys import argv

root = Path(__file__).parent.parent.parent.as_uri().replace("file:///", "")
if root[0] != sep and root[1] != ':':
    root = sep + root

def build(goos, goarch, exe):
    filename = split(exe)[1]
    with ZipFile(f"westpunk_{goos}.zip", "w") as zip:
        environ["GOOS"] = goos
        environ["GOARCH"] = goarch
        run(["go", "install"])
        exepath = join(root, "bin", exe)
        thispath = exepath.replace(join("bin", exe), join("pkg", "westpunk"))
        copy(exepath, join(thispath, filename))
        zip.write(join(thispath, filename), arcname=filename)
        zip.write(join(thispath, "background.png"), arcname="background.png")
        zip.write(join(thispath, "database.db"), arcname="database.db")
        zip.write(join(thispath, "dirtlayerground.png"), arcname="dirtlayerground.png")
        zip.write(join(thispath, "grasslayerground.png"), arcname="grasslayerground.png")
        zip.write(join(thispath, "oaklog.png"), arcname="oaklog.png")
        zip.write(join(thispath, "spritesheet.png"), arcname="spritesheet.png")
        zip.write(join(thispath, "tree.png"), arcname="tree.png")
        remove(join(thispath, filename))

if __name__ == "__main__" and len(argv) > 1:
    build(argv[1], argv[2], argv[3])