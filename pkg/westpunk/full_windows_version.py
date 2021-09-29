from subprocess import run
from zipfile import ZipFile
from pathlib import Path
from shutil import copy
from os import remove

root = Path(__file__).parent.parent.parent.as_uri()

if __name__ == "__main__":
    with ZipFile("westpunk_windows.zip", "w") as zip:
        run("go install")
        exepath = (root+"/bin/westpunk.exe").replace("file:///", "").replace("/", "\\")
        thispath = exepath.replace("\\bin\\westpunk.exe", "\\pkg\\westpunk\\")
        copy(exepath, thispath+"westpunk.exe")
        zip.write(thispath+"westpunk.exe", arcname="westpunk.exe")
        zip.write(thispath+"background.png", arcname="background.png")
        zip.write(thispath+"database.db", arcname="database.db")
        zip.write(thispath+"dirtlayerground.png", arcname="dirtlayerground.png")
        zip.write(thispath+"grasslayerground.png", arcname="grasslayerground.png")
        zip.write(thispath+"oaklog.png", arcname="oaklog.png")
        zip.write(thispath+"spritesheet.png", arcname="spritesheet.png")
        zip.write(thispath+"tree.png", arcname="tree.png")
        remove(thispath+"westpunk.exe")