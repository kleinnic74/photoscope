#!/bin/sh

FILES=$(find bin/arm -type f)
DEST_DIR=photoscope/bin

TARGET=admin@192.168.1.200

ssh ${TARGET} mkdir -p ${DEST_DIR}
scp $FILES ${TARGET}:photoscope/bin/

