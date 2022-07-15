# aozora2fmt

A command line tool for converting Aozora Bunko
([青空文庫](https://www.aozora.gr.jp/index.html)) files to better
formats.

## Description

This tool walks through the text and replaces the following:
* JIS codepoint markers are replaced with the correct UTF-8 character
* Ruby text markers are replaced with tags appropriate for the output format
* Headers are replaced with markers appropriate for the output format
* Page breaks are replaced with a marker appropriate for the output format
* The info block at the start of the file is removed (unless `[-d]` is specified)

## Installation

Simply clone the repository and run:

	make install
