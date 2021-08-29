#!/usr/bin/env python3

import os
import subprocess

DIR = "."

for file in os.listdir(DIR):
    if file.endswith(".dot"):
        svg_file = file.replace(".dot", ".svg")
        subprocess.run(["dot", "-Tsvgz", "-o", svg_file, file])