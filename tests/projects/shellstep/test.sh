#!/bin/bash

# This is intended to be called from wercker/test-all.sh, which sets the required environment variables
# if you run this file directly, you need to set $wercker, $workingDir and $testDir
# as a convenience, if these are not set then assume we're running from the local directory 
if [ -z ${wercker} ]; then wercker=$PWD/../../../wercker; fi
if [ -z ${workingDir} ]; then workingDir=$PWD/../../../.werckertests; mkdir -p "$workingDir"; fi
if [ -z ${testsDir} ]; then testsDir=$PWD/..; fi

testShellstep () {
  testName=shellstep
  testDir=$testsDir/shellstep
  printf "testing %s... " "$testName"

  # the shellstep test passes locally but fails (with inappropriate ioctl) when run in wercker
  if [ -n ${WERCKER_ROOT} ]; then 
    # Running in wercker, shellstep test would not work so don't set --enable-dev-steps which will cause the step to be skipped (which will still allow the test to pass)
    enable-dev-steps=
  else 
    # Running locally, shellstep test will work so set --enable-dev-steps to allow the step to be executed (which will allow the test to pass)
    enable-dev-steps=--enable-dev-steps
  fi
  
  $wercker build "$testDir" ${enable-dev-steps} --docker-local --working-dir "$workingDir" &> "${workingDir}/${testName}.log"
  if [ $? -ne 0 ]; then
    printf "failed\n"
    if [ "${workingDir}/${testName}.log" ]; then
      cat "${workingDir}/${testName}.log"
    fi
    return 1
  fi

  printf "passed\n"
  return 0

}

testShellstep || exit 1
