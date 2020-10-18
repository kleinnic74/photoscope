#!/bin/sh

files=$(find bin/arm -type f)

scp $files admin@192.168.1.200:.

