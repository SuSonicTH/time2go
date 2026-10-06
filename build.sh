#!/bin/bash
go build -ldflags "-s -w" -trimpath && upx --ultra-brute --lzma time2go.exe
