#!/bin/bash

BASEDIR=$(dirname "$0")
cd $BASEDIR/../frontend
prettier --write --ignore-path  .prettierignore --config .prettierrc