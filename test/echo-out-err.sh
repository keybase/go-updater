#!/usr/bin/env bash

echo "stdout output"
(>&2 echo "stderr output")
